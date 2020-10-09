#!/usr/bin/env python3

import argparse
import csv
import datetime
import subprocess
import sys

def getrow(d):
    args = ['/usr/bin/ledger', 'bal', 'Assets', 'Liabilities', '-X', '$', '--real', '--end', str(d + datetime.timedelta(days=1))]
    p = subprocess.run(args, stdout=subprocess.PIPE, check=True, universal_newlines=True)
    o = p.stdout.split('\n')[-2].strip()
    return [str(d), o]

def dothings(start_date, end_date):
    if end_date < start_date:
        raise ValueError('end date {end_date} < start date {start_date}'.format(end_date=end_date, start_date=start_date))

    w = csv.writer(sys.stdout)
    w.writerow(['Date', 'Net Worth'])

    cur_date = start_date
    while True:
        if end_date < cur_date:
            break
        w.writerow(getrow(cur_date))
        cur_date = cur_date + datetime.timedelta(days=1)

def todate(s):
    invalidError = '{} is not in YYYY-MM-DD format'.format(s)
    ymd = s.strip().split('-')
    if len(ymd) != 3 or len(ymd[0]) != 4 or len(ymd[1]) != 2 or len(ymd[2]) != 2:
        raise ValueError(invalidError)
    return datetime.date(int(ymd[0]), int(ymd[1]), int(ymd[2]))

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('start_date', help='Start date in YYYY-MM-DD format.')
    parser.add_argument('end_date', help='End date in YYYY-MM-DD format.')
    args = parser.parse_args()
    dothings(todate(args.start_date), todate(args.end_date))

if __name__ == '__main__':
    main()
