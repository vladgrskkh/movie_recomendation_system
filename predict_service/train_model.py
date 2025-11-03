import pandas as pd
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import joblib
import ast

from models.recommender_model import MovieRecommender

df = pd.read_csv("models/data/movies_dataset.csv")

def parse_features(x):
    try:
        data = ast.literal_eval(x)
        if isinstance(data, list):
            return " ".join([d["name"] for d in data])
    except:
        return ""
    return ""

df["genres"] = df["genres"].fillna("[]").apply(parse_features)
df["keywords"] = df["keywords"].fillna("[]").apply(parse_features)
df["overview"] = df["overview"].fillna("")

df["text_features"] = df["overview"] + " " + df["genres"] + " " + df["keywords"]

vectorizer = TfidfVectorizer(stop_words="english", max_features=10000)
tfidf_matrix = vectorizer.fit_transform(df["text_features"])

similarity = cosine_similarity(tfidf_matrix)

model = MovieRecommender(df[["title"]], similarity)
joblib.dump(model, "predict_service/models/recommender.pkl")
print("TMDB-based recommender model saved to models/recommender.pkl")
