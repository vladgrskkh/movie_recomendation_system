import numpy as np
from sklearn.metrics.pairwise import cosine_similarity


def cosine_similarities(
    query_vector,
    matrix
) -> np.ndarray:
    """
    Возвращает cosine similarity между query и всей матрицей
    """
    sims = cosine_similarity(query_vector, matrix)
    return sims.flatten()
