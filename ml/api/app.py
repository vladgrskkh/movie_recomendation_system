from fastapi import FastAPI, HTTPException
import pandas as pd
from fastapi import Query
from ml.api.dependencies import recommender_service
from ml.api.schemas import (
    RecommendationRequest,
    RecommendationResponse,
    MovieSearchResponse
)
from ml.api.dependencies import recommender_service

app = FastAPI(
    title="Movie Recommendation API",
    version="1.0.0"
)


@app.post(
    "/recommend",
    response_model=RecommendationResponse
)
def recommend_movies(request: RecommendationRequest):

    if not request.movie_ids:
        raise HTTPException(
            status_code=400,
            detail="movie_ids must not be empty"
        )

    try:
        results_df = recommender_service.recommender.recommend_from_multiple(
            movie_ids=request.movie_ids,
            top_k=request.top_k
        )
    except ValueError as e:
        raise HTTPException(
            status_code=400,
            detail=str(e)
        )

    recommendations = [
        {
            "movie_id": int(row.movie_id),
            "title": row.title,
            "score": float(row.score)
        }
        for _, row in results_df.iterrows()
    ]

    return {"recommendations": recommendations}

@app.get(
    "/search",
    response_model=MovieSearchResponse
)
def search_movies(
    q: str = Query(..., min_length=2, description="Movie title"),
    limit: int = Query(10, ge=1, le=50)
):
    movies = recommender_service.recommender.movies_df

    mask = movies["title"].str.contains(
        q,
        case=False,
        na=False
    )

    results = (
        movies[mask]
        .sort_values("release_year", ascending=False)
        .head(limit)
    )

    return {
        "results": [
            {
                "movie_id": int(row.movie_id),
                "title": row.title,
                "release_year": (
                    int(row.release_year)
                    if not pd.isna(row.release_year)
                    else None
                )
            }
            for _, row in results.iterrows()
        ]
    }

