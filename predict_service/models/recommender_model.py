class MovieRecommender:
    def __init__(self, df, similarity):
        self.df = df
        self.similarity = similarity

    def recommend(self, movie_title, top_k=5):
        if movie_title not in self.df["title"].values:
            return [("Movie not found", 0.0)]

        idx = self.df[self.df["title"] == movie_title].index[0]
        sim_scores = list(enumerate(self.similarity[idx]))
        sim_scores = sorted(sim_scores, key=lambda x: x[1], reverse=True)[1:top_k+1]
        recs = [(self.df.iloc[i]["title"], float(score)) for i, score in sim_scores]
        return recs