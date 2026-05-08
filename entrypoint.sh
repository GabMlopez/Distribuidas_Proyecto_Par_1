#!/bin/sh

echo "🚀 Iniciando servidor de chat distribuido..."

# Esperar a que MongoDB esté listo (solo para desarrollo local)
if [ ! -z "$WAIT_FOR_DB" ]; then
    echo "⏳ Esperando a que MongoDB esté listo..."
    while ! nc -z mongodb 27017; do
        sleep 1
    done
    echo "✅ MongoDB está listo!"
fi

# Crear bucket en MinIO si no existe
if [ ! -z "$MINIO_ENDPOINT" ]; then
    echo "📦 Verificando bucket de MinIO..."
    # Usar mc (minio client) para crear bucket
    wget -q https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x mc
    ./mc alias set myminio http://$MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
    ./mc mb myminio/$MINIO_BUCKET --ignore-existing
    ./mc anonymous set download myminio/$MINIO_BUCKET
    rm ./mc
    echo "✅ Bucket listo!"
fi

# Ejecutar migrations si es necesario
# echo "🔄 Ejecutando migrations..."
# ./migrate

echo "✅ Servidor listo! Iniciando aplicación..."
# Iniciar la aplicación
exec ./main