create table photos
(
	id integer
		primary key
		 autoincrement,
	hash varchar(100),
	name varchar(255),
	description varchar(255),
	camera varchar(255),
	lens varchar(255),
	focal_length integer,
	iso integer,
	shutter_speed varchar(255),
	aperture real,
	time_viewed bigint,
	rating real,
	category integer,
	location varchar(255),
	privacy bool,
	latitude real,
	longitude real,
	taken_at datetime,
	width integer,
	height integer,
	nsfw bool,
	licence_type integer,
	url varchar(255)
)
;

create unique index uix_photos_hash
	on photos (hash)
;

create table users
(
	id integer
		primary key
		 autoincrement,
	username varchar(255),
	firstname varchar(255),
	lastname varchar(255),
	gender integer,
	email varchar(255),
	address varchar(255),
	city varchar(255),
	state varchar(255),
	zip varchar(255),
	country varchar(255),
	about varchar(255),
	locale varchar(255),
	show_nsfw bool,
	user_url varchar(255),
	admin bool,
	avatar_url varchar(255),
	api_key varchar(255),
	uid varchar(255),
	gid varchar(255),
	home_dir varchar(255),
	password varchar(255)
)
;

create unique index uix_users_email
	on users (email)
;

create unique index uix_users_username
	on users (username)
;

