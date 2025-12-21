import pandas as pd
from pathlib import Path
from build_csv import OUTPUT_DIR as DATA_DIR


SCHEMAS = {
    "movies.csv": {
        "required_columns": [
            "movie_id", "title", "release_year", "runtime",
            "budget", "revenue", "popularity",
            "vote_average", "vote_count",
            "original_language", "director_id"
        ],
        "unique": ["movie_id"],
        "not_null": ["movie_id", "title"]
    },

    "movie_genres.csv": {
        "required_columns": ["movie_id", "genre_id"],
        "not_null": ["movie_id", "genre_id"]
    },

    "movie_keywords.csv": {
        "required_columns": ["movie_id", "keyword_id"],
        "not_null": ["movie_id", "keyword_id"]
    },

    "movie_cast.csv": {
        "required_columns": ["movie_id", "person_id", "cast_order"],
        "not_null": ["movie_id", "person_id"]
    },

    "movie_music.csv": {
        "required_columns": ["movie_id", "person_id", "role"],
        "not_null": ["movie_id", "person_id", "role"]
    },

    "movie_countries.csv": {
        "required_columns": ["movie_id", "country_code"],
        "not_null": ["movie_id", "country_code"]
    },

    "movie_overviews.csv": {
        "required_columns": ["movie_id", "overview"],
        "not_null": ["movie_id", "overview"]
    },
}



def load_csv(path: Path) -> pd.DataFrame:
    if not path.exists():
        raise FileNotFoundError(f"Missing file: {path}")
    df = pd.read_csv(path)
    if df.empty:
        raise ValueError(f"Empty file: {path}")
    return df


def check_columns(df, required_columns, name):
    missing = set(required_columns) - set(df.columns)
    if missing:
        raise ValueError(f"{name}: missing columns {missing}")


def check_not_null(df, columns, name):
    nulls = df[columns].isnull().sum()
    bad = nulls[nulls > 0]
    if not bad.empty:
        raise ValueError(f"{name}: null values found:\n{bad}")


def check_unique(df, columns, name):
    if df.duplicated(subset=columns).any():
        raise ValueError(f"{name}: duplicate values in {columns}")
    
def check_movie_id_integrity(df, movie_ids, name):
    bad_ids = set(df["movie_id"]) - movie_ids
    if bad_ids:
        raise ValueError(
            f"{name}: contains unknown movie_id(s): {list(bad_ids)[:10]}"
        )

def validate_all_csv(data_dir: Path):
    print("🔍 Validating CSV files...\n")

    movies_path = data_dir / "movies.csv"
    movies_df = load_csv(movies_path)
    movie_ids = set(movies_df["movie_id"])

    for file_name, schema in SCHEMAS.items():
        path = data_dir / file_name
        print(f"→ Checking {file_name}")

        df = load_csv(path)

        check_columns(df, schema["required_columns"], file_name)

        if "not_null" in schema:
            check_not_null(df, schema["not_null"], file_name)

        if "unique" in schema:
            check_unique(df, schema["unique"], file_name)

        if file_name != "movies.csv":
            check_movie_id_integrity(df, movie_ids, file_name)

        print(
            f"   rows={len(df)}, "
            f"unique movies={df['movie_id'].nunique()}"
        )

    print("\n✅ All CSV files passed validation")

def log(dir):
    df = pd.read_csv(dir)
    print(df[df["country_code"].isna()])

if __name__ == "__main__":
    validate_all_csv(DATA_DIR)
   