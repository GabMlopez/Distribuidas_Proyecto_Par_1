
chmod +x entrypoint.sh

echo "🏗️  Construyendo imágenes Docker..."
docker build -t chat-backend:latest .

echo "🧪 Verificando build..."
docker run --rm chat-backend:latest ./main --help

echo "✅ Setup completado!"