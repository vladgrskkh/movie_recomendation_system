import pandas as pd
from typing import Tuple
from scipy.sparse import csr_matrix


def build_multi_hot(
    df: pd.DataFrame,
    index_col: str,
    feature_col: str,
    top_k: int | None = None
) -> Tuple[csr_matrix, list]:
    """
    multi-hot encoder.
    Returns:
    matrix: csr_matrix (n_movies, n_features)
    feature_ids: list of feature values (columns)
    """

    if top_k is not None:
        top_features = (
            df[feature_col]
            .value_counts()
            .head(top_k)
            .index
        )
        df = df[df[feature_col].isin(top_features)]

    pivot = (
        df
        .assign(value=1)
        .pivot_table(
            index=index_col,
            columns=feature_col,
            values="value",
            fill_value=0
        )
    )

    return csr_matrix(pivot.values), list(pivot.columns), pivot.index
