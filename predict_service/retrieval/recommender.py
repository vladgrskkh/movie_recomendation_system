import numpy as np
from .similarity import cosine_similarities
from sklearn.preprocessing import normalize
from scipy.sparse import csr_matrix

class ContentRecommender:
    def __init__(self, movie_ids, movie_matrix, movies_df):
        self.movie_ids = movie_ids
        self.movie_matrix = movie_matrix
        self.movies_df = movies_df

        self.id_to_idx = {
            mid: i for i, mid in enumerate(movie_ids)
        }

    def recommend_similar(
        self,
        movie_id: int,
        top_k: int = 10
    ):
        """
        Рекомендации на основе одного фильма
        """
        if movie_id not in self.id_to_idx:
            raise ValueError("Unknown movie_id")

        idx = self.id_to_idx[movie_id]
        query = self.movie_matrix[idx]

        sims = cosine_similarities(query, self.movie_matrix)
        sims[idx] = -1

        top_idx = np.argsort(sims)[-top_k:][::-1]

        results = self.movies_df.iloc[top_idx][
            ["movie_id", "title"]
        ].copy()

        results["score"] = sims[top_idx]
        return results
    
    def recommend_from_multiple(
        self,
        movie_ids: list[int],
        top_k: int = 10,
        weights: list[float] | None = None
    ):
        """
        Рекомендации на основе нескольких фильмов
        """

        indices = []
        for mid in movie_ids:
            if mid in self.id_to_idx:
                indices.append(self.id_to_idx[mid])

        if not indices:
            raise ValueError("No valid movie_ids provided")

        vectors = self.movie_matrix[indices]

        # aggregation
        if weights is not None:
            w = np.array(weights).reshape(-1, 1)
            query = vectors.multiply(w).sum(axis=0)
        else:
            query = vectors.mean(axis=0)

        # np.matrix -> csr_matrix
        query = csr_matrix(query)

        # ---- normalize  ----
        query = normalize(query)

        sims = cosine_similarities(query, self.movie_matrix)

        # исключаем сами фильмы
        for idx in indices:
            sims[idx] = -1

        top_idx = np.argsort(sims)[-top_k:][::-1]

        results = self.movies_df.iloc[top_idx][
            ["movie_id", "title"]
        ].copy()

        results["score"] = sims[top_idx]
        return results







    