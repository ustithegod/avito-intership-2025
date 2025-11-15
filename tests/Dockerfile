FROM python:3.12-slim

WORKDIR /project

# Установим зависимости для тестов
COPY tests/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt

# По умолчанию гоняем все тесты в /project/tests
CMD ["pytest", "-ra", "-vv", "/project/tests"]
