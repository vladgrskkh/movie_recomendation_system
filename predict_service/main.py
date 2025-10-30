import grpc
from concurrent import futures
import joblib
from v1.predict import predict_pb2, predict_pb2_grpc

from models.recommender_model import MovieRecommender

class Recommendation(predict_pb2_grpc.RecommendationServicer):
    def __init__(self):
        print("Loading model")
        self.model = joblib.load("models/recommender.pkl")
        print("Model loaded")

    def Recommend(self, request, context):
        movie_title = request.movieTitle
        recs = self.model.recommend(movie_title)
        recommendations = [
            predict_pb2.Recommendation(title=title, score=score)
            for title, score in recs
        ]
        return predict_pb2.RecommendResponse(recommendations=recommendations)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    predict_pb2_grpc.add_RecommendationServicer_to_server(
        Recommendation(), server
    )
    server.add_insecure_port("[::]:50051")
    print("gRPC Movie Recommender running on port 50051")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
