
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

	function dumpsome(bindata) {
		return "<<binary data: len: " + bindata.length + ">>";
	}

	exports.Fcall.prototype.asString = function() {
		switch(this.Type) {
		case mt.Tversion:
			return sprintf("Tversion tag %d msize %d version '%s'", this.Tag, this.Msize, this.Version);
		case mt.Rversion:
			return sprintf("Rversion tag %d msize %d version '%s'", this.Tag, this.Msize, this.Version);
		case Tauth:
			return sprintf("Tauth tag %d afid %d uname %s aname %s",
				this.Tag, this.Afid, this.Uname, this.Aname);
		case Rauth:
			return sprintf("Rauth tag %d qid %v", this.Tag, this.Qid);
		case Tattach:
			return sprintf("Tattach tag %d fid %d afid %d uname %s aname %s",
				this.Tag, this.Fid, this.Afid, this.Uname, this.Aname);
		case Rattach:
			return sprintf("Rattach tag %d qid %v", this.Tag, this.Qid);
		case Rerror:
			return sprintf("Rerror tag %d ename %s", this.Tag, this.Ename);
		case Tflush:
			return sprintf("Tflush tag %d oldtag %d", this.Tag, this.Oldtag);
		case Rflush:
			return sprintf("Rflush tag %d", this.Tag);
		case Twalk:
			return sprintf("Twalk tag %d fid %d newfid %d wname %v",
				this.Tag, this.Fid, this.Newfid, this.Wname);
		case Rwalk:
			return sprintf("Rwalk tag %d wqid %v", this.Tag, this.Wqid);
		case Topen:
			return sprintf("Topen tag %d fid %d mode %d", this.Tag, this.Fid, this.Mode);
		case Ropen:
			return sprintf("Ropen tag %d qid %v iouint %d", this.Tag, this.Qid, this.Iounit);
		case Tcreate:
			return sprintf("Tcreate tag %d fid %d name %s perm %v mode %d",
				this.Tag, this.Fid, this.Name, this.Perm, this.Mode);
		case Rcreate:
			return sprintf("Rcreate tag %d qid %v iouint %d", this.Tag, this.Qid, this.Iounit);
		case Tread:
			return sprintf("Tread tag %d fid %d offset %d count %d",
				this.Tag, this.Fid, this.Offset, this.Count);
		case Rread:
			return sprintf("Rread tag %d count %d %s",
				this.Tag, len(this.Data), dumpsome(this.Data));
		case Twrite:
			return sprintf("Twrite tag %d fid %d offset %d count %d %s",
				this.Tag, this.Fid, this.Offset, len(this.Data), dumpsome(this.Data));
		case Rwrite:
			return sprintf("Rwrite tag %d count %d", this.Tag, this.Count);
		case Tclunk:
			return sprintf("Tclunk tag %d fid %d", this.Tag, this.Fid);
		case Rclunk:
			return sprintf("Rclunk tag %d", this.Tag);
		case Tremove:
			return sprintf("Tremove tag %d fid %d", this.Tag, this.Fid);
		case Rremove:
			return sprintf("Rremove tag %d", this.Tag);
		case Tstat:
			return sprintf("Tstat tag %d fid %d", this.Tag, this.Fid);
			/*
			 * some hand work needed here
		case Rstat:
			d, err := UnmarshalDir(f.Stat)
			if err == nil {
				return fmt.Sprintf("Rstat tag %d stat(%d bytes)",
					f.Tag, len(f.Stat))
			}
			return fmt.Sprintf("Rstat tag %d stat %v", f.Tag, d)
		case Twstat:
			d, err := UnmarshalDir(f.Stat)
			if err == nil {
				return fmt.Sprintf("Twstat tag %d fid %d stat(%d bytes)",
					f.Tag, f.Fid, len(f.Stat))
			}
			return fmt.Sprintf("Twstat tag %d fid %d stat %v", f.Tag, f.Fid, d)
		case Rwstat:
			return fmt.Sprintf("FidRwstat tag %d", f.Tag)
			*/
		default:
			throw "Invalid type";
		}
	};
