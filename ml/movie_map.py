import matplotlib.pyplot as plt
from vectorization import signature_matrix, movies_df
from sklearn.manifold import TSNE

# Уменьшаем до 2D
tsne = TSNE(n_components=2, random_state=42, perplexity=30, metric='cosine')
embedding = tsne.fit_transform(signature_matrix)

movies_df['x'] = embedding[:, 0]
movies_df['y'] = embedding[:, 1]

# Берём главный жанр фильма (первый в списке) для цвета
movies_df['main_genre'] = movies_df['genres'].apply(
    lambda x: x[0]['name'] if isinstance(x, list) and len(x) > 0 else "Unknown"
)

plt.figure(figsize=(12, 8))
for genre in movies_df['main_genre'].unique():
    subset = movies_df[movies_df['main_genre'] == genre]
    plt.scatter(subset['x'], subset['y'], label=genre, alpha=0.6, s=30)

plt.legend(markerscale=2, bbox_to_anchor=(1.05, 1), loc='upper left')
plt.title("Карта фильмов (t-SNE, жанры+год+рейтинг)")
plt.xlabel("X")
plt.ylabel("Y")
plt.show()
