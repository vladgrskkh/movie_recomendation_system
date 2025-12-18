import pandas as pd
from pathlib import Path
from ml.features import MovieVectorBuilder
from ml.retrieval.recommender import ContentRecommender
from pathlib import Path
import pandas as pd

DATA_DIR = Path("ml/data/movies/ru/csv")
CPI_PATH = Path("ml/data/external/cpi.csv")

movies = pd.read_csv(DATA_DIR / "movies.csv")

builder = MovieVectorBuilder(DATA_DIR, CPI_PATH)
movie_ids, movie_matrix = builder.build()

movies_df = pd.read_csv(DATA_DIR / "movies.csv")

recommender = ContentRecommender(
    movie_ids,
    movie_matrix,
    movies_df
)

# Toy Story, Finding Nemo, Shrek
seed_movies = [862, 12, 808]

print(
    recommender.recommend_from_multiple(
        movie_ids=seed_movies,
        top_k=10
    )
)