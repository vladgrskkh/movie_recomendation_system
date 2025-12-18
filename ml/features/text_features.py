import pandas as pd
from typing import Dict, Tuple
from sklearn.feature_extraction.text import TfidfVectorizer
from scipy.sparse import csr_matrix


def build_text_features(
    data_dir,
    max_features: int = 5000,
    ngram_range: Tuple[int, int] = (1, 2),
):
    """
    Returns:
    - text_matrix: csr_matrix (n_movies_with_overview, n_features)
    - movie_id_to_row: Dict[movie_id -> row_index]
    - vectorizer: fitted TfidfVectorizer
    """

    overviews = pd.read_csv(data_dir / "movie_overviews.csv")

    # базовая чистка
    overviews["overview"] = (
        overviews["overview"]
        .astype(str)
        .str.strip()
    )

    overviews = overviews[overviews["overview"] != ""]

    vectorizer = TfidfVectorizer(
        max_features=max_features,
        ngram_range=ngram_range,
        min_df=5,        
        max_df=0.9       
    )

    text_matrix = vectorizer.fit_transform(overviews["overview"])

    movie_id_to_row: Dict[int, int] = dict(
        zip(overviews["movie_id"], range(len(overviews)))
    )

    return (
        csr_matrix(text_matrix),
        movie_id_to_row,
        vectorizer
    )
