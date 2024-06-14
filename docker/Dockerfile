# Usar a imagem base oficial do Golang
FROM golang:1.16

# Definir o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copiar os arquivos go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código-fonte para o contêiner
COPY . .

# Construir o executável
RUN go build -o main .

# Executar o executável
CMD ["./main"]