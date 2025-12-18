from pathlib import Path
from ml.features import MovieVectorBuilder

DATA_DIR = Path("ml/data/movies/ru/csv")
CPI_PATH = Path("ml/data/external/cpi.csv")

builder = MovieVectorBuilder(DATA_DIR, CPI_PATH)
movie_ids, movie_matrix = builder.build()

print(movie_matrix.shape)
print(movie_matrix[0].nnz)