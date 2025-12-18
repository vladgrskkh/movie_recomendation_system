from pathlib import Path
import pandas as pd

from ml.features import MovieVectorBuilder
from ml.retrieval.recommender import ContentRecommender

class RecommenderService:
    def __init__(self):
        DATA_DIR = Path("ml/data/movies/ru/csv")
        CPI_PATH = Path("ml/data/external/cpi.csv")

        builder = MovieVectorBuilder(DATA_DIR, CPI_PATH)
        movie_ids, movie_matrix = builder.build()

        movies_df = pd.read_csv(DATA_DIR / "movies.csv")

        self.recommender = ContentRecommender(
            movie_ids=movie_ids,
            movie_matrix=movie_matrix,
            movies_df=movies_df
        )


# singleton
recommender_service = RecommenderService()
