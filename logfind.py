#!/usr/bin/python3
import datetime
from os.path import join
import sys
from typing import List, Tuple
import os
import re

def yesno(q: str) -> bool:
    while True:
        ans = input(f'{q} [Y/N]: ')
        if ans.lower() in ('y', 'yes'):
            return True
        elif ans.lower() in ('n', 'no'):
            return False

# [chomp:server] 2022/04/28 10:52:09 Starting chomp at Thu Apr 28 10:52:09
# Finds log files for given date.
def find_files(date: datetime.date) -> List[Tuple[str, datetime.datetime]]:
    # note: we dont look at chomp.log or chomp.log.old.
    if not os.path.isdir('logs'):
        return []

    files = []
    for file in os.listdir('logs'):
        path = join('logs', file)
        with open(path, 'r') as f:
            line = f.readline()
            datepart = re.search('\\[chomp:server\\] (.+) Starting chomp at', line).group(1)
            logdate = datetime.datetime.strptime(datepart, "%Y/%m/%d %H:%M:%S")

            if logdate.date() == date:
                files.append((path, logdate))

    return files


def prompt_results(results: List[Tuple[str, datetime.datetime]]):
    print(f'Found {len(results)} log file(s) matching your search.')
    if len(results) == 0:
        return []

    filter = yesno('Do you want to filter by time?')

    if filter:
        logs = []
        spec_min = True
        while True:
            try:
                time = input('Input the time you want to filter by (HH:[MM]): ')
                if time[len(time) - 1] == ':':
                    fmt = "%H:"
                    spec_min = False
                else:
                    fmt = "%H:%M"

                filterby = datetime.datetime.strptime(time, fmt)
                break
            except:
                pass

        for res in results:
            if res[1].time().hour == filterby.hour and (not spec_min or res[1].time().minute == filterby.minute):
                logs.append(res)

        return logs
    else:
        return results


def main():
    if len(sys.argv) == 2:
        try:
            filterby  = datetime.datetime.strptime(sys.argv[1], "%Y-%m-%d")
        except Exception as e:
            print(f'Error: {e}')
            return
    else:
        while True:
            try:
                date = input('Input the date you want to find logs for (YYYY-MM-DD): ')
                filterby = datetime.datetime.strptime(date, "%Y-%m-%d")
                break
            except:
                pass


    results = prompt_results(find_files(filterby.date()))
    if len(results) == 0:
        print('No results matched the given criteria.')
        return
    else:
        print(f'Found {len(results)} results for your criteria:')
        for res in results:
            print(f'File: {res[0]}')
            print(f'\tWritten at {res[1]}')

if __name__ == '__main__':
    main()

