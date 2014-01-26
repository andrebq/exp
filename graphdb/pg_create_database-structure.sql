-- drop the functions
drop function if exists fn_keyword(p_keyword char varying(255));
drop function if exists fn_keyword_code (p_code integer);
drop function if exists fn_new_node(p_kind integer);
drop function if exists fn_new_edge(p_source bigint, p_target bigint, p_kind integer);
drop function if exists fn_node_data(p_node bigint, p_attribute integer, p_data char varying);
drop function if exists fn_fetch_node_with_data(p_nodeid bigint);

-- drop the views
drop view if exists vw_node_with_contents;

-- drop the keyword index
drop table if exists keywords;

-- drop old content
drop table if exists nodecontents;
drop table if exists edgecontents;

-- drop graph tables
drop table if exists edges;
drop table if exists nodes;

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

create index idx_keywords_word on keywords(keyword);
create index idx_keywords_code on keywords(keycode);

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
	kind integer not null,
	contents char varying,

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
	kind integer not null,
	contents char varying,

	constraint pk_edgecontents
		primary key (edgeid),

	constraint fk_edge
		foreign key (edgeid)
		references edges(edgeid)
) with ( oids = false );

alter table edgecontents owner to graphdb;

create view vw_node_with_contents as
	select n.nodeid, n.kind, nc.kind attr, nc.contents
	from nodes n
		inner join nodecontents nc
			on nc.nodeid = n.nodeid;

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

create function fn_keyword_code(p_keycode integer) returns char varying as $BODY$
declare
	v_keyword char varying;
begin
	select keyword into v_keyword
	from keywords
	where keycode = p_keycode;

	return v_keyword;
end;
$BODY$ language plpgsql;

alter function fn_keyword_code(integer) owner to graphdb;

create function fn_new_node (p_kind integer) returns bigint as $BODY$
declare
	p_nodeid bigint;
begin
	if (select not exists(select 1 from keywords where keycode = p_kind)) then
		raise exception 'invalid node kind. it must be a valid keyword';
	end if;

	select into p_nodeid nextval('seq_nodes');

	insert into nodes(nodeid, kind)
	values (p_nodeid, p_kind);

	return p_nodeid;
end;
$BODY$ language plpgsql;

create function fn_new_edge (p_source bigint, p_target bigint, p_kind integer) returns bigint as $BODY$
declare
	p_edgeid bigint;
begin
	if (select not exists(select 1 from keywords where keycode = p_kind)) then
		raise exception 'invalid node kind. it must be a valid keyword';
	end if;

	select into p_edgeid nextval('seq_edges');

	insert into edges(edgeid, sourceid, targetid, kind)
	values (p_edgeid, p_start, p_target, p_kind);

	return p_edgeid;
end;
$BODY$ language plpgsql;

create function fn_node_data(p_node bigint, p_attribute integer, p_data char varying) returns integer as $BODY$
declare
begin
	if (select not exists(select 1 from keywords where keycode = p_attribute)) then
		raise exception 'invalid attribute kind. it must be a valid keyword';
	end if;

	update nodecontents
	set contents = p_data
	where nodeid = p_node
		and kind = p_attribute;

	if found then
		raise notice 'achou';
		return 1;
	else
		insert into nodecontents(nodeid, kind, contents)
		values (p_node, p_attribute, p_data);
		return 2;
	end if;

	return null;
end;
$BODY$ language plpgsql;

create function fn_fetch_node_with_data(p_nodeid bigint) returns setof vw_node_with_contents as $BODY$
begin
	return query select * from vw_node_with_contents where nodeid = p_nodeid;
	return;
end;
$BODY$ language plpgsql;

select fn_keyword(':core/node_name');
select fn_keyword(':core/edge_name');

begin;

	do $BODY$
	declare
		v_rootid bigint;
		v_datacode integer;
	begin
		select into v_rootid fn_new_node(fn_keyword(':core/rootnode'));
		select into v_datacode fn_node_data(v_rootid, fn_keyword(':core/metadata/created_at'), cast(current_timestamp as char varying));
	end; $BODY$ language plpgsql;

commit;
