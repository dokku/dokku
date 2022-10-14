FROM python:3.10.8-buster

RUN pip install --user pipenv
ENV PATH="${PATH}:/root/.local/bin"
ENV PIPENV_VENV_IN_PROJECT=0

COPY . /app

WORKDIR /app

RUN ls -lah

RUN rm -rf .venv

RUN mkdir .venv

RUN pipenv install

CMD pipenv run gunicorn -b 0.0.0.0:8000 --workers=3 --threads=5 project.wsgi:application
