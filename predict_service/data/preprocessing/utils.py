import pandas as pd
import json
from pathlib import Path
from typing import List, Dict


def load_tmdb_json(path: str | Path) -> List[Dict]:
    path = Path(path)

    if not path.exists():
        raise FileNotFoundError(f"JSON file not found: {path}")

    with path.open("r", encoding="utf-8") as f:
        data = json.load(f)

    if not isinstance(data, list):
        raise ValueError("TMDB JSON must be a list of movie objects")

    return data


def save_csv(rows, path, columns=None):
    if not rows:
        print(f"[WARN] No data to save: {path.name}")
        return

    df = pd.DataFrame(rows)

    if columns:
        df = df[columns]

    print(f"[OK] Saved {len(df)} rows → {path}") 

    df.to_csv(path, index=False)



def ensure_dir(path: str | Path) -> Path:
    path = Path(path)
    path.mkdir(parents=True, exist_ok=True)
    return path
