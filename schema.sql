-- Players

create table players (
    id int primary key,
    name text not null
);

create index players_name_idx on players (name);
create index players_lower_name_idx on players (lower(name) text_pattern_ops);

-- Alliances

create table alliances (
    tag  text primary key,
    name text not null
);

create index alliances_lower_tag_idx on alliances (lower(tag) text_pattern_ops);
create index alliances_lower_name_idx on alliances (lower(name) text_pattern_ops);

-- Alliance members

create table alliance_members (
    alliance_tag text not null references alliances,
    player_id int not null references players,
    updated timestamp not null,
    primary key (alliance_tag, player_id, updated)
);

-- Player points

create function create_player_points(kind text) returns void as $$
begin
    execute '
        create table player_points_' || kind || ' (
            player_id int not null references players,
            updated timestamp not null,
            points int check (points >= 0) not null,
            primary key (player_id, updated)
        )';
end
$$ language plpgsql;

select create_player_points('points');
select create_player_points('fleet');
select create_player_points('research');

drop function create_player_points(kind text);

-- Player points week

create function create_player_points_week(kind text) returns void as $$
begin
    execute '
        create materialized view player_points_' || kind || '_week as (
            select player_id, updated, points
              from player_points_' || kind || '
             where updated >= (
                       select max(updated) - interval ''1 week''
                         from player_points_' || kind || '
                   )
        )';
    execute 'create index on player_points_' || kind || '_week (player_id)';
end
$$ language plpgsql;

select create_player_points_week('points');
select create_player_points_week('fleet');
select create_player_points_week('research');

drop function create_player_points_week(kind text);

-- Player points month

create function create_player_points_month(kind text) returns void as $$
begin
    execute '
        create materialized view player_points_' || kind || '_month as (
            select player_id, updated, points
              from player_points_' || kind || '
             where updated >= (
                       select max(updated) - interval ''1 month''
                         from player_points_' || kind || '
                        where extract(hours from updated) = 0
                   ) and
                   extract(hours from updated) = 0
        )';
    execute 'create index on player_points_' || kind || '_month (player_id)';
end
$$ language plpgsql;

select create_player_points_month('points');
select create_player_points_month('fleet');
select create_player_points_month('research');

drop function create_player_points_month(kind text);

-- Player points all time

create function create_player_points_all_time(kind text) returns void as $$
begin
    execute '
        create materialized view player_points_' || kind || '_all_time as (
            select player_id, updated, points
              from player_points_' || kind || '
             where extract(isodow from updated) = 1 and
                   extract(hours from updated) = 0
        )';
    execute 'create index on player_points_' || kind || '_all_time (player_id)';
end
$$ language plpgsql;

select create_player_points_all_time('points');
select create_player_points_all_time('fleet');
select create_player_points_all_time('research');

drop function create_player_points_all_time(kind text);

-- Player ranking

create function create_player_ranking(kind text, period text) returns void as $$
begin
    execute '
        create materialized view player_ranking_' || kind || '_' || period || ' as (
            select player_id,
                   updated,
                   (rank() over (partition by updated order by points desc))::int as rank
              from player_points_' || kind || '_' || period || '
        )';
    execute 'create index on player_ranking_' || kind || '_' || period || ' (player_id)';
end
$$ language plpgsql;

select create_player_ranking('points', 'week');
select create_player_ranking('points', 'month');
select create_player_ranking('points', 'all_time');
select create_player_ranking('fleet', 'week');
select create_player_ranking('fleet', 'month');
select create_player_ranking('fleet', 'all_time');
select create_player_ranking('research', 'week');
select create_player_ranking('research', 'month');
select create_player_ranking('research', 'all_time');

drop function create_player_ranking(kind text, period text);

-- Alliance points

create function create_alliance_points(kind text, period text) returns void as $$
begin
    execute '
        create materialized view alliance_points_' || kind || '_' || period || ' as (
              select am.alliance_tag,
                     p.updated,
                     sum(p.points)::int as points
                from player_points_' || kind || '_' || period || ' p
                join alliance_members am
                  on am.player_id = p.player_id and
                     am.updated = p.updated
            group by am.alliance_tag, p.updated
        )';
    execute 'create index on alliance_points_' || kind || '_' || period || ' (alliance_tag)';
end
$$ language plpgsql;

select create_alliance_points('points', 'week');
select create_alliance_points('points', 'month');
select create_alliance_points('points', 'all_time');
select create_alliance_points('fleet', 'week');
select create_alliance_points('fleet', 'month');
select create_alliance_points('fleet', 'all_time');
select create_alliance_points('research', 'week');
select create_alliance_points('research', 'month');
select create_alliance_points('research', 'all_time');

drop function create_alliance_points(kind text, period text);

-- Alliance ranking

create function create_alliance_ranking(kind text, period text) returns void as $$
begin
    execute '
        create materialized view alliance_ranking_' || kind || '_' || period || ' as (
              select alliance_tag,
                     updated,
                     (rank() over (partition by updated order by points desc))::int as rank
                from alliance_points_' || kind || '_' || period || '
        )';
    execute 'create index on alliance_ranking_' || kind || '_' || period || ' (alliance_tag)';
end
$$ language plpgsql;

select create_alliance_ranking('points', 'week');
select create_alliance_ranking('points', 'month');
select create_alliance_ranking('points', 'all_time');
select create_alliance_ranking('fleet', 'week');
select create_alliance_ranking('fleet', 'month');
select create_alliance_ranking('fleet', 'all_time');
select create_alliance_ranking('research', 'week');
select create_alliance_ranking('research', 'month');
select create_alliance_ranking('research', 'all_time');

drop function create_alliance_ranking(kind text, period text);

-- Top

create function create_top(kind text) returns void as $$
begin
    execute '
        create materialized view top_' || kind || ' as (
                     select p.name,
                            am.alliance_tag,
                            pts1.points as points,
                            (rank() over (order by pts1.points desc))::int as rank,
                            pts1.points - pts2.points as week_difference,
                            (rank() over (order by pts1.points - pts2.points desc nulls last))::int as week_difference_rank,
                            pts1.points - pts3.points as month_difference,
                            (rank() over (order by pts1.points - pts3.points desc nulls last))::int as month_difference_rank
                       from players p
                       join player_points_' || kind || ' pts1
                         on pts1.player_id = p.id and
                            pts1.updated = (select max(updated) from player_points_' || kind || ')
            left outer join player_points_' || kind || ' pts2
                         on pts2.player_id = p.id and
                            pts2.updated = pts1.updated - interval ''1 week''
            left outer join player_points_' || kind || ' pts3
                         on pts3.player_id = p.id and
                            pts3.updated = pts1.updated - interval ''1 month''
            left outer join alliance_members am
                         on am.player_id = p.id and
                            am.updated = pts1.updated
        )';
    execute 'create index on top_' || kind || ' (points)';
    execute 'create index on top_' || kind || ' (week_difference)';
    execute 'create index on top_' || kind || ' (month_difference)';
end
$$ language plpgsql;

select create_top('points');
select create_top('fleet');
-- select create_top('research');

drop function create_top(kind text);
