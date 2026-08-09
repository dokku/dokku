import sys
import time


def main(args):
    # a log shipper attaches to a container after it receives the start event,
    # and a cron container is removed the moment it exits. sleep on either side
    # of the output so that the shipper has time to attach before the line is
    # written and time to read it before the container is gone.
    time.sleep(5)
    print(' '.join(args[1:]), flush=True)
    time.sleep(3)


if __name__ == '__main__':
    main(sys.argv)
