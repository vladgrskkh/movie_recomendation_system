from pydantic import BaseModel
from typing import List, Optional


class RecommendationRequest(BaseModel):
    movie_ids: List[int]
    top_k: Optional[int] = 10


class MovieRecommendation(BaseModel):
    movie_id: int
    title: str
    score: float


class RecommendationResponse(BaseModel):
    recommendations: List[MovieRecommendation]

class MovieSearchResult(BaseModel):
    movie_id: int
    title: str
    release_year: int | None


class MovieSearchResponse(BaseModel):
    results: List[MovieSearchResult]
