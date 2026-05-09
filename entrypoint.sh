
echo "🚀 Iniciando servidor de chat distribuido..."
if [ ! -z "$WAIT_FOR_DB" ]; then
    echo "⏳ Esperando a que MongoDB esté listo..."
    while ! nc -z mongodb 27017; do
        sleep 1
    done
    echo "✅ MongoDB está listo!"
fi

if [ ! -z "$MINIO_ENDPOINT" ]; then
    echo "📦 Verificando bucket de MinIO..."
    wget -q https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x mc
    ./mc alias set myminio http://$MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
    ./mc mb myminio/$MINIO_BUCKET --ignore-existing
    ./mc anonymous set download myminio/$MINIO_BUCKET
    rm ./mc
    echo "✅ Bucket listo!"
fi


echo "✅ Servidor listo! Iniciando aplicación..."
exec ./main