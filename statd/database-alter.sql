-- Run this after you called initdb
--
-- most of the code here, create indexes on the system
ALTER TABLE buckets
  ADD CONSTRAINT pk_buckets PRIMARY KEY(id);

ALTER TABLE buckets
  OWNER TO statsd;

-- Index: buckets_bucketname_idx

-- DROP INDEX buckets_bucketname_idx;

CREATE INDEX buckets_bucketname_idx
  ON buckets
  USING btree
  (bucket COLLATE pg_catalog."default");

-- Index: buckets_servertime_index

-- DROP INDEX buckets_servertime_index;

CREATE INDEX buckets_servertime_index
  ON buckets
  USING btree
  (servertime);


