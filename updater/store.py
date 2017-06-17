import sys

import psycopg2

import config

def update_players(cursor, entries):
    print('updating players')
    sql = """
        INSERT INTO players (id, name) VALUES (%(id)s, %(name)s)
        ON CONFLICT (id) DO UPDATE SET name = %(name)s
    """
    cursor.executemany(sql, entries)

def update_alliances(cursor, entries):
    print('updating alliances')
    sql = """
        INSERT INTO alliances (tag, name) VALUES (%(tag)s, %(name)s)
        ON CONFLICT (tag) DO UPDATE SET name = %(name)s
    """
    # FIXME: Remove duplicates
    entries = filter(lambda entry: entry['alliance_tag'] is not None, entries)
    records = [{'tag': entry['alliance_tag'], 'name': entry['alliance_name']} for entry in entries]
    cursor.executemany(sql, records)

def update_alliance_members(cursor, time, entries):
    print('updating alliance members')
    entries = list(filter(lambda entry: entry['alliance_tag'] is not None, entries))
    sql = 'INSERT INTO alliance_members (alliance_tag, player_id, updated) VALUES {}'.format(
            ','.join(['%s'] * len(entries)))
    records = [(entry['alliance_tag'], entry['id'], time) for entry in entries]
    cursor.execute(sql, records)

def update_player_specific_points(cursor, kind, time, entries):
    print('updating player points', kind)
    sql = 'INSERT INTO player_points_{} (player_id, updated, points) VALUES {}'.format(
            kind, ','.join(['%s'] * len(entries)))
    records = [(entry['id'], time, entry['points']) for entry in entries]
    cursor.execute(sql, records)

def update_player_points(cursor, time, points, fleet, research):
    print('updating player points')
    update_player_specific_points(cursor, 'points', time, points)
    update_player_specific_points(cursor, 'fleet', time, fleet)
    update_player_specific_points(cursor, 'research', time, research)

def refresh_points_and_ranking(cursor):
    # alliance must be after player
    for who in ['player', 'alliance']:
        # ranking must be after points
        for what in ['points', 'ranking']:
            for kind in ['points', 'fleet', 'research']:
                for period in ['week', 'month', 'all_time']:
                    print('refreshing {} {} {} {}'.format(who, what, kind, period))
                    sql = 'REFRESH MATERIALIZED VIEW {}_{}_{}_{}'.format(
                            who, what, kind, period)
                    cursor.execute(sql)

def refresh_top(cursor, kind):
    print('refreshing top ' + kind)
    cursor.execute('REFRESH MATERIALIZED VIEW top_' + kind)

def update_all(time, points, fleet, research):
    with psycopg2.connect('dbname=' + config.dbname) as connection:
        with connection.cursor() as cursor:
            # The order matters
            update_players(cursor, points)
            update_alliances(cursor, points)
            update_alliance_members(cursor, time, points)
            update_player_points(cursor, time, points, fleet, research)
            refresh_points_and_ranking(cursor)
            refresh_top('points')
            refresh_top('fleet')
