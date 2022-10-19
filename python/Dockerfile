FROM python:alpine

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY . .

USER 10050

EXPOSE 9090

CMD ["python", "main.py"]
