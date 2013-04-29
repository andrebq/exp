
	/**
	 * Struct used to represent 9p messages
	 * */
	exports.Fcall = function() {

		this.Type    = 0;//uint8
		this.Fid     = 0;//uint32
		this.Tag     = 0;//uint16
		this.Msize   = 0;//uint32
		this.Version = 0;//string   // Tversion, Rversion
		this.Oldtag  = 0;//uint16   // Tflush
		this.Ename   = 0;//string   // Rerror
		this.Qid     = null;// Rattach, Ropen, Rcreate
		this.Iounit  = 0;//uint32   // Ropen, Rcreate
		this.Aqid    = null;//Qid      // Rauth
		this.Afid    = 0;//uint32   // Tauth, Tattach
		this.Uname   = 0;//string   // Tauth, Tattach
		this.Aname   = 0;//string   // Tauth, Tattach
		this.Perm    = 0;//Perm     // Tcreate
		this.Name    = 0;//string   // Tcreate
		this.Mode    = 0;//uint8    // Tcreate, Topen
		this.Newfid  = 0;//uint32   // Twalk
		this.Wname   = null;//[]string // Twalk
		this.Wqid    = null;//[]Qid    // Rwalk
		this.Offset  = 0;//uint64   // Tread, Twrite
		this.Count   = 0;//uint32   // Tread, Rwrite
		this.Data    = null;//[]byte   // Twrite, Rread
		this.Stat    = null;//[]byte   // Twstat, Rstat

		// 9P2000.u extensions
		this.Errno     = 0;//uint32 // Rerror
		this.Uid       = 0;//uint32 // Tattach, Tauth
		this.Extension = 0;//string // Tcreate

	};

	mt = {
		Tversion : 100,
		Rversion : 101,
		Tauth : 102,
		Rauth : 103,
		Tattach : 104,
		Rattach : 105,
		//Terror // illegal
		Rerror : 106,
		Tflush : 107,
		Rflush : 108,
		Twalk : 109,
		Rwalk : 110,
		Topen : 111,
		Ropen : 112,
		Tcreate : 113,
		Rcreate : 114,
		Tread : 115,
		Rread : 116,
		Twrite : 117,
		Rwrite : 118,
		Tclunk : 119,
		Rclunk : 120,
		Tremove : 121,
		Rremove : 122,
		Tstat : 123,
		Rstat : 124,
		Twstat : 125,
		Rwstat : 126,
		Tmax : 127
	};
	exports.MessageType = mt;

	exports.Fcall.prototype.asString = function() {
		switch(this.Type) {
		case mt.Tversion:
			return sprintf("Tversion tag %d msize %d version '%s'", this.Tag, this.Msize, this.Version);
		case mt.Rversion:
			return sprintf("Rversion tag %d msize %d version '%s'", this.Tag, this.Msize, this.Version);
		default:
			throw "Invalid type";
		}
	};
