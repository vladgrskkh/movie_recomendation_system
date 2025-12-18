from typing import List, Dict, Optional


# ---------- Director ----------

def extract_director_id(crew: List[Dict]) -> Optional[int]:
    for person in crew:
        if person.get("job") == "Director":
            return int(person.get("id"))
    return None


# ---------- Cast (top N actors) ----------

def extract_top_cast(cast: List[Dict], n: int = 7) -> List[Dict]:
    cast = [
        c for c in cast
        if c.get("known_for_department") == "Acting"
    ]

    cast = sorted(cast, key=lambda x: x.get("order", 999))
    return cast[:n]


# ---------- Production Countries ----------

def extract_production_countries(
    movie_id: int,
    production_countries: List[Dict]
) -> List[Dict]:
    rows = []

    if not production_countries:
        return [{
            "movie_id": movie_id,
            "country_code": "UNK"
        }]

    for country in production_countries:
        code = country.get("iso_3166_1")

        # --- отбрасываем NaN ---
        if isinstance(code, float) and math.isnan(code):
            continue

        # --- валидный ISO-код ---
        if isinstance(code, str):
            code = code.strip().upper()

            # строго 2 символа A-Z и НЕ "NA"
            if len(code) == 2 and code.isalpha() and code != "NA":
                rows.append({
                    "movie_id": movie_id,
                    "country_code": code
                })

    if not rows:
        rows.append({
            "movie_id": movie_id,
            "country_code": "UNK"
        })

    return rows


# ---------- Music ----------

MUSIC_JOBS = {
    "Original Music Composer",
    "Composer",
    "Songs",
    "Music"
}

def extract_music(crew: List[Dict]) -> List[Dict]:
    """
    Extract music-related crew members.
    """
    result = []

    for person in crew:
        if (
            person.get("department") == "Sound"
            and person.get("job") in MUSIC_JOBS
        ):
            result.append({
                "person_id": person.get("id"),
                "role": person.get("job")
            })

    return result

# ---------- Overviews ----------
def extract_overview(movie: Dict) -> Optional[Dict]:
    overview = movie.get("overview")
    if overview and overview.strip():
        return {
            "movie_id": int(movie["id"]),
            "overview": overview.strip()
        }
    return None
