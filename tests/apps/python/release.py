#!/usr/bin/env python
import os


def main():
    print("GLOBAL_SECRET: {0}".format(os.getenv('GLOBAL_SECRET')))
    print("SECRET_KEY: {0}".format(os.getenv('SECRET_KEY')))


if __name__ == '__main__':
    main()
