-- Self-contained fixtures for db-gen integration / e2e tests.
-- Applied by the Go test harness (requireTestDB) before each integration run.
-- Everything lives in an isolated `dbgen_test` schema that is dropped and
-- recreated on every apply, so tests always see a known, clean structure.

drop schema if exists dbgen_test cascade;
create schema dbgen_test;

-- 1. Simple scalar-returning function -------------------------------------
create function dbgen_test.scalar_sum(a int, b int) returns int
	language sql as
$$
select a + b;
$$;

-- 2. Table-returning function (structured model) --------------------------
create function dbgen_test.get_rows()
	returns table(id int, label text)
	language sql as
$$
select 1, 'one'
union all
select 2, null;
$$;

-- 3. Void-returning function ----------------------------------------------
create function dbgen_test.do_nothing(a int) returns void
	language plpgsql as
$$
begin
end;
$$;

-- 4. Procedure ------------------------------------------------------------
create procedure dbgen_test.noop_proc(a int)
	language plpgsql as
$$
begin
end;
$$;

-- 5. Function with a DEFAULT parameter (optional) -------------------------
create function dbgen_test.with_defaults(required_val int, optional_val int default 5) returns int
	language sql as
$$
select required_val + optional_val;
$$;

-- 6. Function with a context parameter (_user_id) -------------------------
create function dbgen_test.with_context(_user_id int, name text) returns int
	language sql as
$$
select _user_id;
$$;

-- 7. Overloaded function (requires MappedName in config) ------------------
create function dbgen_test.overloaded(x int) returns int
	language sql as
$$
select x;
$$;

create function dbgen_test.overloaded(x text) returns text
	language sql as
$$
select x;
$$;

-- 8. Copy-target staging table -------------------------------------------
-- created_by / job_run_id are injected context columns; the rest are data.
create table dbgen_test.copy_target_demo
(
	created_by text   not null,
	job_run_id bigint not null,
	company_id text   not null,
	name       text,
	amount     text
);

-- 9. Function with sensitive parameters (drives db-gen security levels) ----
-- Exercises every SecurityLevel resolution path in one signature:
--   _user_id     -> context parameter (mapped in the e2e config)
--   username     -> none    (global ParameterSecurityMappings)
--   display_name -> secure  (falls through to DefaultParameterSecurityLevel)
--   email        -> omit    (per-function override beats the default)
--   password     -> strict  (global)
--   api_token    -> strict  (global)
--   secret_note  -> omit    (global) and optional+nullable (has a DEFAULT)
create function dbgen_test.register_user(
	_user_id     int,
	username     text,
	display_name text,
	email        text,
	password     text,
	api_token    text,
	secret_note  text default null
) returns int
	language sql as
$$
select _user_id;
$$;
