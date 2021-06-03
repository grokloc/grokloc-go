// Package schemas contains db schemas
// NOTE: production, migrated schemas are not managed here
package schemas

// exported table names
const (
	OrgsTableName  = "orgs"
	UsersTableName = "users"
)

// AppCreate creates the entire app schema
const AppCreate = `
create table if not exists users (
       api_secret text unique not null,
       api_secret_digest text unique not null,
       id text unique not null,
       display_name text not null,
       display_name_digest text not null,
       email text unique not null,
       email_digest text unique not null,
       org text not null,
       password text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer, 
       primary key (id));
-- STMT
create unique index if not exists email_org on users (email_digest, org);
-- STMT
create trigger if not exists users_ctime_trigger after insert on users
begin
        update users set 
        ctime = strftime('%s','now'), 
        mtime = strftime('%s','now') 
        where id = new.id;
end;
-- STMT
create trigger if not exists users_mtime_trigger after update on users
begin
        update users set mtime = strftime('%s','now') 
        where id = new.id;
end;
-- STMT
create table if not exists orgs (
       id text unique not null,
       name text unique not null,
       owner text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer,
       primary key (id));
-- STMT
create trigger if not exists orgs_ctime_trigger after insert on orgs
begin
        update orgs set 
        ctime = strftime('%s','now'), 
        mtime = strftime('%s','now') 
        where id = new.id;
end;
-- STMT
create trigger if not exists orgs_mtime_trigger after update on orgs
begin
        update orgs set mtime = strftime('%s','now') 
        where id = new.id;
end;
-- STMT
create table if not exists repositories (
       id text unique not null,
       name text unique not null,
       org text not null,
       path text not null,
       url text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer,
       primary key (id));
-- STMT
create trigger if not exists repositories_ctime_trigger after insert on repositories
begin
        update repositories set 
        ctime = strftime('%s','now'), 
        mtime = strftime('%s','now') 
        where id = new.id;
end;
-- STMT
create trigger if not exists repositories_mtime_trigger after update on repositories
begin
        update repositories set mtime = strftime('%s','now') 
        where id = new.id;
end;
`
