FROM python:3.13.9-alpine
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY server.py ws.py Procfile ./
EXPOSE 5000
CMD ["python", "server.py", "web"]
