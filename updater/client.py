import datetime
import http.cookiejar
import re
import sys
import urllib.parse
import urllib.request

import lxml.html

import config

cj = http.cookiejar.CookieJar()
opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
session = ''

def make_url(path):
    global session
    url = 'http://' + config.domain + '/' + path
    if session != '':
        url += '&session=' + session
    return url

def login(username, password):
    global session
    url = make_url('game/reg/login2.php')
    params = {
        'login': username,
        'pass': password,
        'Abschicken': 'Login',
        'v': '2',
    }
    data = urllib.parse.urlencode(params).encode('ascii')
    with opener.open(url, data) as f:
        content = f.read().decode('utf-8')
        if 'wrong password' in content:
            print('login failed: invalid username or password')
            sys.exit(2)
        m = re.search(r'session=([0-9a-f]{12})\&', content)
        session = m.group(1)
        print('logging successful')

def update_time():
    print('fetching update time')
    params = {
        'page': 'stat',
        'who': 'player',
        'type': 'pts',
        'start': 1
    }
    url = make_url('game/index.php?' + urllib.parse.urlencode(params))
    with opener.open(url) as f:
        content = f.read().decode('utf-8')
        root = lxml.html.fromstring(content)
        trs = root.xpath('//table[@width="519"]/tr')
        updated = datetime.datetime.strptime(trs[0][0].text, 'Statistics (Updated: %Y-%m-%d, %H:%M)')
        print(updated)
        return updated

def fetch_single_page(kind, page):
    start = page * 100 + 1
    end = start + 99
    print('fetching', kind, start, '-', end)
    params = {
        'page': 'stat',
        'who': 'player',
        'type': kind,
        'start': start
    }
    url = make_url('game/index.php?' + urllib.parse.urlencode(params))
    with opener.open(url) as f:
        content = f.read().decode('utf-8')
        root = lxml.html.fromstring(content)
        entries = []
        trs = root.xpath('//table[@width="519"]/tr')
        for tr in trs[3:]:
            try:
                ths = list(tr)
                entry = {}
                # Player ID
                entry['id'] = int(re.search(r'messageziel=(\d+)', ths[2][0].attrib['href']).group(1))
                # Player name
                entry['name'] = ths[1].text.strip()
                # Player points
                entry['points'] = int(ths[4].text.strip().replace('.', ''))
                # Player rank
                entry['rank'] = int(ths[0].text.strip())
                # Alliance tag
                m = re.search(r'allytag=(.*)', ths[3][0].attrib['href'])
                if m is not None:
                    entry['alliance_tag'] = urllib.parse.unquote_plus(m.group(1))
                else:
                    entry['alliance_tag'] = None
                # Alliance name
                entry['alliance_name'] = ths[3][0].text.strip()
                if entry['alliance_name'] == '':
                    entry['alliance_name'] = None
                entries.append(entry)
            except AttributeError:
                raise
        if len(entries) == 0 or entries[0]['rank'] != start:
            return []
        else:
            return entries

def fetch_all(kind):
    page = 0
    stats = []
    while True:
        s = fetch_single_page(kind, page)
        if s == []:
            break
        stats.extend(s)
        page += 1
    return stats

def points():
    return fetch_all('pts')

def fleet():
    return fetch_all('flt')

def research():
    return fetch_all('res')
