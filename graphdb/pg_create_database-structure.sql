create extension if not exists hstore;

-- drop the keyword index
drop table if exists keywords;

-- drop old content
drop table if exists nodecontents;
drop table if exists edgecontents;

-- drop graph tables
drop table if exists edges;
drop table if exists nodes;

-- drop the functions
drop function if exists fn_keyword(p_keyword char varying(255));

drop sequence if exists seq_keywords;
drop sequence if exists seq_nodes;
drop sequence if exists seq_edges;

create sequence seq_keywords
	increment by 1
	no minvalue
	no maxvalue
	cache 10
	cycle;

create sequence seq_nodes
	increment by 1
	no minvalue
	no maxvalue
	cache 10
	cycle;

create sequence seq_edges
	increment by 1
	no minvalue
	no maxvalue
	cache 10
	cycle;

create table if not exists keywords (
	keyword char varying(255),
	keycode integer not null,
	constraint pk_keywords
		primary key(keyword),
	constraint unq_code
		unique (keycode)
) with ( oids = false );

alter table keywords owner to graphdb;


create table if not exists nodes (
	nodeid bigint,
	kind integer,

	constraint pk_nodes
		primary key(nodeid)
) with ( oids = false );

alter table nodes owner to graphdb;

create table if not exists nodecontents (
	nodeid bigint,
	keycode int not null,
	contents bytea,

	constraint pk_nodecontents
		primary key(nodeid),

	constraint fk_node
		foreign key (nodeid)
		references nodes(nodeid)
) with ( oids = false );

alter table nodecontents owner to graphdb;

create table if not exists edges (
	edgeid bigint,
	kind integer not null,
	sourceid bigint not null,
	targetid bigint not null,

	constraint pk_edges
		primary key (edgeid),

	constraint fk_source_node
		foreign key (sourceid)
		references nodes(nodeid),

	constraint fk_target_node
		foreign key (targetid)
		references nodes(nodeid)
) with ( oids = false );

alter table edges owner to graphdb;

create table if not exists edgecontents (
	edgeid bigint not null,
	keycode int not null,
	contents bytea,

	constraint pk_edgecontents
		primary key (edgeid),

	constraint fk_edge
		foreign key (edgeid)
		references edges(edgeid)
) with ( oids = false );

alter table edgecontents owner to graphdb;

create function fn_keyword (p_keyword char varying(255)) returns integer as $BODY$
declare
	v_keycode integer;
begin
	select keycode into v_keycode
	from keywords
	where keyword = p_keyword;

	if not found then
		select into v_keycode nextval('seq_keywords');
		insert into keywords (keyword, keycode)
		values (p_keyword, v_keycode);
	end if;

	return v_keycode;
end;
$BODY$ language plpgsql;

alter function fn_keyword(char varying) owner to graphdb;

select fn_keyword(':core/node_name');
select fn_keyword(':core/edge_name');