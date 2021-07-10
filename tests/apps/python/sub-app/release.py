#!/usr/bin/env python
import os


def main():
    print("SECRET_KEYS: {0}".format(os.getenv('SECRET_KEY')))


if __name__ == '__main__':
    main()
