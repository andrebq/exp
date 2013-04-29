
	exports.VERSION9P = "9P2000";
	exports.MAXWELEM  = 16;

	exports.OREAD     = 0;
	exports.OWRITE    = 1;
	exports.ORDWR     = 2;
	exports.OEXEC     = 3;
	exports.OTRUNC    = 16;
	exports.OCEXEC    = 32;
	exports.ORCLOSE   = 64;
	exports.ODIRECT   = 128;
	exports.ONONBLOCK = 256;
	exports.OEXCL     = 0x1000;
	exports.OLOCK     = 0x2000;
	exports.OAPPEND   = 0x4000;

	exports.AEXIST = 0;
	exports.AEXEC  = 1;
	exports.AWRITE = 2;
	exports.AREAD  = 4;

	exports.QTDIR     = 0x80;
	exports.QTAPPEND  = 0x40;
	exports.QTEXCL    = 0x20;
	exports.QTMOUNT   = 0x10;
	exports.QTAUTH    = 0x08;
	exports.QTTMP     = 0x04;
	exports.QTSYMLINK = 0x02;
	exports.QTFILE    = 0x00;

	exports.DMDIR       = 0x80000000;
	exports.DMAPPEND    = 0x40000000;
	exports.DMEXCL      = 0x20000000;
	exports.DMMOUNT     = 0x10000000;
	exports.DMAUTH      = 0x08000000;
	exports.DMTMP       = 0x04000000;
	exports.DMSYMLINK   = 0x02000000;
	exports.DMDEVICE    = 0x00800000;
	exports.DMNAMEDPIPE = 0x00200000;
	exports.DMSOCKET    = 0x00100000;
	exports.DMSETUID    = 0x00080000;
	exports.DMSETGID    = 0x00040000;
	exports.DMREAD      = 0x4;
	exports.DMWRITE     = 0x2;
	exports.DMEXEC      = 0x1;

	exports.NOTAG   = 0xffff;
	exports.NOFID   = 0xffffffff;
	exports.NOUID   = 0xffffffff;
	exports.IOHDRSZ = 24;

