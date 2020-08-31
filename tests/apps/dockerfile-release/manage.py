#!/usr/bin/env python
"""Django's command-line utility for administrative tasks."""
import os


def main():
    print("SECRET_KEY: {0}".format(os.getenv('SECRET_KEY')))


if __name__ == '__main__':
    main()
