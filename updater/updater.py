import sys

import client
import store

def main():
    if len(sys.argv) != 3:
        print('Usage:', sys.argv[0], '<username> <password>')
        sys.exit(1)
    client.login(sys.argv[1], sys.argv[2])
    updated = client.update_time()
    points = client.points()
    fleet = client.fleet()
    research = client.research()
    updated2 = client.update_time()
    if updated != updated2:
        print('statistics updated during fetching data')
        sys.exit(3)
    store.update_all(updated, points, fleet, research)

if __name__ == '__main__':
    main()
