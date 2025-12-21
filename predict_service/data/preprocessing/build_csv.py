import pandas as pd
from pathlib import Path
from utils import load_tmdb_json, ensure_dir, save_csv
from extractors import (
    extract_director_id,
    extract_overview,
    extract_top_cast,
    extract_production_countries,
    extract_music,
)

# ---------- Paths ----------

BASE_DIR = Path(__file__).resolve().parents[2]

DATA_DIR = BASE_DIR / "data" / "movies" / "ru"
OUTPUT_DIR = DATA_DIR / "csv"
INPUT_FILE = DATA_DIR / "json" / "tmdb_movies_ru_10_000.json"


# ---------- Main preprocessing ----------

def main():
    print("Loading TMDB JSON...")
    movies_raw = load_tmdb_json(INPUT_FILE)
    print(f"Loaded {len(movies_raw)} movies")

    ensure_dir(OUTPUT_DIR)

    # Containers
    movies_rows = []
    genres_rows = []
    keywords_rows = []
    cast_rows = []
    countries_rows = []
    music_rows = []
    overviews_rows = []

    for movie in movies_raw:
        movie_id = movie.get("id")
        if movie_id is None:
            continue

        release_year = None
        release_date = movie.get("release_date")
        if release_date:
            try:
                release_year = int(release_date[:4])
            except ValueError:
                pass

        movies_rows.append({
            "movie_id": movie_id,
            "title": movie.get("title"),
            "release_year": release_year,
            "runtime": movie.get("runtime") or None,
            "budget": movie.get("budget") or None,
            "revenue": movie.get("revenue") or None,
            "popularity": movie.get("popularity"),
            "vote_average": movie.get("vote_average"),
            "vote_count": movie.get("vote_count"),
            "original_language": movie.get("original_language"),
            "director_id": extract_director_id(
                movie.get("credits", {}).get("crew", [])
            )
        })

        for genre in movie.get("genres", []):
            genres_rows.append({
                "movie_id": movie_id,
                "genre_id": genre.get("id")
            })

        for kw in movie.get("keywords", {}).get("keywords", []):
            keywords_rows.append({
                "movie_id": movie_id,
                "keyword_id": kw.get("id")
            })

        countries_rows.extend(
            extract_production_countries(
                movie_id,
                movie.get("production_countries", [])
            )
        )

        top_cast = extract_top_cast(
            movie.get("credits", {}).get("cast", []),
            n=7
        )

        for actor in top_cast:
            cast_rows.append({
                "movie_id": movie_id,
                "person_id": actor.get("id"),
                "cast_order": actor.get("order", 999)
            })

        music_people = extract_music(
            movie.get("credits", {}).get("crew", [])
        )

        for m in music_people:
            music_rows.append({
                "movie_id": movie_id,
                "person_id": m["person_id"],
                "role": m["role"]
            })
            
        overview_row = extract_overview(movie)
        if overview_row:
            overviews_rows.append(overview_row)
    

    df_movies = pd.DataFrame(movies_rows)

    df_movies["movie_id"] = df_movies["movie_id"].astype("Int64")
    df_movies["director_id"] = df_movies["director_id"].astype("Int64")
    df_movies["release_year"] = df_movies["release_year"].astype("Int64")
    df_movies["runtime"] = df_movies["runtime"].astype("Int64")
    df_movies["budget"] = df_movies["budget"].astype("Int64")
    df_movies["revenue"] = df_movies["revenue"].astype("Int64")
    df_movies["vote_count"] = df_movies["vote_count"].astype("Int64")

    df_movies["popularity"] = df_movies["popularity"].astype("float32")
    df_movies["vote_average"] = df_movies["vote_average"].astype("float32")

    df_movies.to_csv(OUTPUT_DIR / "movies.csv", index=False)

    print("Saving CSV files...")

    save_csv(genres_rows, OUTPUT_DIR / "movie_genres.csv")
    save_csv(keywords_rows, OUTPUT_DIR / "movie_keywords.csv")
    save_csv(cast_rows, OUTPUT_DIR / "movie_cast.csv")
    save_csv(countries_rows, OUTPUT_DIR / "movie_countries.csv")
    save_csv(music_rows, OUTPUT_DIR / "movie_music.csv")
    save_csv(overviews_rows, OUTPUT_DIR / "movie_overviews.csv")
    print("Done.")
    print(f"CSV files saved to: {OUTPUT_DIR}")


if __name__ == "__main__":
    main()
    