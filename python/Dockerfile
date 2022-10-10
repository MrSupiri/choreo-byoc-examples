FROM python:alpine

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY . .

USER 10016

EXPOSE 5000

CMD ["python", "main.py"]
