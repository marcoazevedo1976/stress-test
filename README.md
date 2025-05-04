# Desafio Stress Test | Pós Go Expert

## Instruções
1. Clone o repositório
2. Entre na pasta do repositório
3. Para gerar a imagem docker execute: 
    ```bash
    docker build -t stress-test .
    ```
4. Para executar a imagem docker execute:
    ```bash
    docker run stress-test --url=http://google.com --requests=1000 --concurrency=10
    ```

OBS: Não foi implementado compose.yaml pois o desafio pede para que o aplicativo seja executado com docker run.