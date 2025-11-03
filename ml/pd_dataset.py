import json
import pandas as pd


# Загружаем json-файл
file_path = "data/en/tmdb_movies.json"
with open(file_path, "r", encoding="utf-8") as f:
    data = json.load(f)

# Превратим список фильмов в DataFrame
movies_df = pd.json_normalize(
    data,
    sep="_",
    max_level=1  # пока только верхний уровень
)

# Создадим отдельные таблицы для cast и crew
cast_records = []
crew_records = []

for movie in data:
    movie_id = movie.get("id")
    credits = movie.get("credits", {})
    
    # cast
    for c in credits.get("cast", []):
        cast_records.append({
            "movie_id": movie_id,
            "cast_id": c.get("cast_id"),
            "person_id": c.get("id"),
            "name": c.get("name"),
            "character": c.get("character"),
            "order": c.get("order")
        })
    
    # crew
    for c in credits.get("crew", []):
        crew_records.append({
            "movie_id": movie_id,
            "person_id": c.get("id"),
            "name": c.get("name"),
            "department": c.get("department"),
            "job": c.get("job")
        })

cast_df = pd.DataFrame(cast_records)
crew_df = pd.DataFrame(crew_records)

# print(f"Создан DataFrame movies_df с размером {movies_df.shape}")
# print(f"Создан DataFrame cast_df с размером {cast_df.shape}")
# print(f"Создан DataFrame crew_df с размером {crew_df.shape}")

