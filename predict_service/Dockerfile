FROM python:3.11-slim AS build

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    gfortran \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --prefix=/install --no-cache-dir -r requirements.txt

COPY . .

FROM python:3.11-slim AS runtime

WORKDIR /app

COPY --from=build /install /usr/local

# Copy application code (excluding unnecessary files)
COPY --from=build /app/common ./common
COPY --from=build /app/models ./models
COPY --from=build /app/v1 ./v1
COPY --from=build /app/main.py ./main.py

EXPOSE 50051

# Default command
CMD ["python", "main.py"]
