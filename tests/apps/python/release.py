#!/usr/bin/env python
import os


def main():
    print("GLOBAL_SECRET: {0}".format(os.getenv('GLOBAL_SECRET')))
    print("SECRET_KEY: {0}".format(os.getenv('SECRET_KEY')))
    with open("/app/.env", "r") as f:
        for line in f.readlines():
            if "DOTENV_KEY" in line:
                print(line)

if __name__ == '__main__':
    main()
