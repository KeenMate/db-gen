drop database if exists db_gen;
create database db_gen;

-- SWITCH TO db_gen database

create schema if not exists "public";
create schema if not exists "ext";

create extension if not exists ltree schema "ext";

create table example_table
(
    number int primary key,
    string text,
    json   json,
    jsonb  jsonb,
    ltree  ext.ltree
);

insert into example_table (number, string, json, jsonb, ltree)
values (1, 'Hello world', json_build_object('key', 'value', 'number', 12),
        json_build_object('key', 'value', 'number', 12)::jsonb, '1.2.3');

create function sum(a int, b int) returns int
    language plpgsql
as
$$
BEGIN
    return a + b;

end;
$$;

create function return_custom_type()
    returns table
            (
                __number int,
                __string text,
                __json   json
            )
    language plpgsql
as
$$
begin

    return query select 1, 'Hello from custom type', json_build_object('key', 'value');
end;
$$;

create or replace function return_setof()
    returns setof example_table
    language plpgsql
as
$$
begin
    return query select * from example_table;
end;
$$;

create function return_void(a int, b int) returns void
    language plpgsql
as
$$
BEGIN
end;

$$;


create procedure procedure(a int, b int)
    language plpgsql
as
$$
begin

end;
$$;

select return_void(1, 2);

create or replace function new_function(name text) returns text
    language plpgsql
as
$$
begin
    return 'Hello ' || name;
end
$$;

select new_function('Honza')
