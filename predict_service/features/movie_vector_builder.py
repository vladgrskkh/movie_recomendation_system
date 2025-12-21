# ml/features/movie_vector_builder.py

import pandas as pd
from pathlib import Path
from scipy.sparse import csr_matrix, hstack, vstack

from .numeric_features import build_numeric_features
from .text_features import build_text_features
from .categorical_features import build_multi_hot


class MovieVectorBuilder:
    def __init__(self, data_dir: Path, cpi_path: Path):
        self.data_dir = Path(data_dir)
        self.cpi_path = Path(cpi_path)

        #signatures movie
        self.movies = None
        self.movie_ids = None
        
        #numeric features
        self.numeric_matrix = None

        #text features
        self.text_matrix = None
        self.text_index = None

        #categorical features
        self.genres_matrix = None
        self.keywords_matrix = None
        self.cat_index = None

    def load_movies(self):
        self.movies = pd.read_csv(self.data_dir / "movies.csv")
        self.movie_ids = self.movies["movie_id"].values

    def build_numeric(self):
        numeric_matrix, _, index = build_numeric_features(
            self.movies,
            cpi_path=str(self.cpi_path)
        )
        assert list(index) == list(self.movie_ids)
        self.numeric_matrix = csr_matrix(numeric_matrix)

    def build_text(self):
        text_matrix, movie_map, _ = build_text_features(self.data_dir)
        self.text_matrix = text_matrix
        self.text_index = movie_map

    def build_categories(self):
        genres = pd.read_csv(self.data_dir / "movie_genres.csv")
        keywords = pd.read_csv(self.data_dir / "movie_keywords.csv")

        # genres
        g_matrix, g_features, g_index = build_multi_hot(
            genres,
            index_col="movie_id",
            feature_col="genre_id"
        )

        # keywords (top-K)
        k_matrix, k_features, k_index = build_multi_hot(
            keywords,
            index_col="movie_id",
            feature_col="keyword_id",
            top_k=3000
        )

        # сохраняем всё отдельно
        self.genres_matrix = g_matrix
        self.genres_index = list(g_index)

        self.keywords_matrix = k_matrix
        self.keywords_index = list(k_index)



    def build(self):
        self.load_movies()
        self.build_numeric()
        self.build_text()
        self.build_categories()

        movie_id_to_pos = {mid: i for i, mid in enumerate(self.movie_ids)}
        genre_id_to_pos = {mid: i for i, mid in enumerate(self.genres_index)}
        keyword_id_to_pos = {mid: i for i, mid in enumerate(self.keywords_index)}

        rows = []

        for movie_id in self.movie_ids:
            parts = []

            # ---- text ----
            if movie_id in self.text_index:
                parts.append(self.text_matrix[self.text_index[movie_id]])
            else:
                parts.append(csr_matrix((1, self.text_matrix.shape[1])))

            # ---- genres ----
            if movie_id in genre_id_to_pos:
                parts.append(self.genres_matrix[genre_id_to_pos[movie_id]] * 0.7)
            else:
                parts.append(csr_matrix((1, self.genres_matrix.shape[1])))

            # ---- keywords ----
            if movie_id in keyword_id_to_pos:
                parts.append(self.keywords_matrix[keyword_id_to_pos[movie_id]] * 0.4)
            else:
                parts.append(csr_matrix((1, self.keywords_matrix.shape[1])))

            # ---- numeric ----
            parts.append(self.numeric_matrix[movie_id_to_pos[movie_id]] * 0.3)

            rows.append(hstack(parts))

        movie_matrix = vstack(rows).tocsr()
        return self.movie_ids, movie_matrix
