from sklearn.preprocessing import MultiLabelBinarizer, MinMaxScaler
from sklearn.metrics.pairwise import cosine_similarity
from pd_dataset import movies_df, cast_df, crew_df
import pandas as pd
import numpy as np

# Жанры (multi-hot encoding)
mlb = MultiLabelBinarizer()
all_genres = set(g['name'] for genres in movies_df['genres'] if isinstance(genres, list) for g in genres)
genre_matrix = mlb.fit_transform(
    movies_df['genres'].apply(lambda x: [g['name'] for g in x] if isinstance(x, list) else [])
)

# Год выхода
movies_df['release_year'] = pd.to_datetime(movies_df['release_date'], errors='coerce').dt.year
movies_df['release_year'] = movies_df['release_year'].fillna(movies_df['release_year'].median())

# # Рейтинг
movies_df['vote_average'].fillna(movies_df['vote_average'].mean())

# Соберём числовые признаки
numeric_features = movies_df[['release_year', 'vote_average']].values

# Нормализация числовых признаков
scaler = MinMaxScaler()
numeric_features_scaled = scaler.fit_transform(numeric_features)

# --- 2. Формируем сигнатуру (вектор признаков) ---
signature_matrix = np.hstack([genre_matrix, numeric_features_scaled])

# --- 3. Считаем косинусное сходство между фильмами ---
similarity_matrix = cosine_similarity(signature_matrix)

# # --- 4. Функция для поиска топ-N похожих фильмов ---
def get_similar_movies(movie_title, top_n=5):
    if movie_title not in movies_df['title'].values:
        return f"Фильм '{movie_title}' не найден в датасете."
    
    idx = movies_df.index[movies_df['title'] == movie_title][0]
    sim_scores = list(enumerate(similarity_matrix[idx]))
    sim_scores = sorted(sim_scores, key=lambda x: x[1], reverse=True)
    
    top_indices = [i for i, score in sim_scores[1:top_n+1]]  # исключаем сам фильм
    return movies_df.iloc[top_indices][['title', 'release_year', 'vote_average']]

# Пример: похожие на "История Игрушек"
print(get_similar_movies("Toy Story", top_n=10))
