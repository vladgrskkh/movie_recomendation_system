import grpc
import pandas as pd
import signal
import os
import multiprocessing
from pythonjsonlogger.json import JsonFormatter
import logging
import sys
from concurrent import futures
from pathlib import Path
from features.movie_vector_builder import MovieVectorBuilder
from retrieval.recommender import ContentRecommender
from v1.predict import predict_pb2, predict_pb2_grpc
from common import types_pb2 as common_pb2

logger = logging.getLogger(__name__)
stdout = logging.StreamHandler(stream=sys.stdout)
jsonFmt = JsonFormatter(
    "%(name)s %(asctime)s %(levelname)s %(filename)s %(lineno)s %(process)d %(message)s",
    rename_fields={"levelname": "severity", "asctime": "timestamp"},
    datefmt="%Y-%m-%dT%H:%M:%SZ",
)
stdout.setFormatter(jsonFmt)
logger.addHandler(stdout)
logger.setLevel(logging.INFO)

class RecommendationService(predict_pb2_grpc.RecommendationServicer):
    def __init__(self):
        data_dir = Path("data/movies/ru/csv")
        cpi_path = Path("data/external/cpi.csv")

        builder = MovieVectorBuilder(data_dir, cpi_path)

        logger.info("Build movie vector")
        movie_ids, movie_matrix = builder.build()

        movies_df = pd.read_csv(data_dir / "movies.csv")
        recommender = ContentRecommender(
            movie_ids=movie_ids,
            movie_matrix=movie_matrix,
            movies_df=movies_df
        )

        self.model = recommender
        logger.info("Model built")        

    def Recommend(self, request, context):
        logger.info(f"Received request for movies: {request.movieID}")
        try:
            recs = self.model.recommend_from_multiple(
                movie_ids=request.movieID,
                top_k=request.topK)
        except ValueError as e:
            logger.error(f"error while recommending: {e}")
            context.abort(
                code=grpc.StatusCode.INVALID_ARGUMENT,
                details=str(e)
            )
    
        recommendations = [
            common_pb2.Recommendation(movieID=row.movie_id, score=row.score)
            for row in recs.itertuples()
        ]
        logger.info(f"Returning recommendations {recs['movie_id']}")
        return predict_pb2.RecommendResponse(recommendations=recommendations)

def serve():
    options = (('grpc.so_reuseport', 1),)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10), options=options)
    predict_pb2_grpc.add_RecommendationServicer_to_server(
        RecommendationService(), server
    )
    server.add_insecure_port("[::]:50051")
    logger.info("gRPC Movie Recommender running on port 50051")

    server.start()
    def handle_sigterm(signum, frame):
        logger.info("gracefully shutting down server")
        done_event = server.stop(30)
        done_event.wait()

    signal.signal(signal.SIGTERM, handle_sigterm)
    signal.signal(signal.SIGINT, handle_sigterm)

    server.wait_for_termination()


def main():
    workers_count = os.cpu_count()
    logger.info(f"Spawning {workers_count} workers")

    workers = []
    for _ in range(workers_count):
        worker = multiprocessing.Process(
            target=serve)
        worker.start()
        workers.append(worker)
    for worker in workers:
        worker.join()

if __name__ == "__main__":
    main()
