import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler


NUMERIC_BASE_COLS = [
    "runtime",
    "vote_average",
    "vote_count",
    "popularity",
    "release_year",
]

ECONOMIC_COLS = [
    "budget",
    "revenue",
]


def load_cpi(cpi_path: str) -> dict:
    """
    year,cpi
    """
    cpi_df = pd.read_csv(cpi_path)
    return dict(zip(cpi_df["year"], cpi_df["cpi"]))


def apply_cpi_adjustment(
    df: pd.DataFrame,
    cpi_map: dict,
    reference_year: int
) -> pd.DataFrame:
    ref_cpi = cpi_map[reference_year]

    for col in ECONOMIC_COLS:
        def adjust(row):
            value = row[col]
            year = row["release_year"]

            if pd.isna(value) or value <= 0:
                return 0.0
            if year not in cpi_map:
                return 0.0

            return value * ref_cpi / cpi_map[year]

        df[f"{col}_real"] = df.apply(adjust, axis=1)

    return df


def build_numeric_features(
    movies_df: pd.DataFrame,
    cpi_path: str):
    """
    Returns:
    numeric_matrix (np.ndarray)
    scaler
    movie_index (pd.Index)
    """

    df = movies_df.copy()

    # ---- Fill missing ----
    df["runtime"] = df["runtime"].fillna(df["runtime"].median())
    df["vote_average"] = df["vote_average"].fillna(df["vote_average"].median())
    df["vote_count"] = df["vote_count"].fillna(0)
    df["popularity"] = df["popularity"].fillna(0)
    df["budget"] = df["budget"].fillna(0)
    df["revenue"] = df["revenue"].fillna(0)
    df["release_year"] = df["release_year"].fillna(df["release_year"].median())

    # ---- CPI ----
    cpi_map = load_cpi(cpi_path)
    reference_year = max(cpi_map.keys())

    df = apply_cpi_adjustment(df, cpi_map, reference_year)

    log_cols = [
        "budget_real",
        "revenue_real",
        "vote_count",
        "popularity",
    ]

    for col in log_cols:
        df[col] = np.log1p(df[col])

    # ---- Select final numeric columns ----
    final_cols = [
        "runtime",
        "budget_real",
        "revenue_real",
        "vote_count",
        "popularity",
        "vote_average",
        "release_year",
    ]

    df_final = df.set_index("movie_id")[final_cols]

    # ---- Scaling ----
    scaler = StandardScaler()
    numeric_matrix = scaler.fit_transform(df_final)

    return numeric_matrix, scaler, df_final.index
