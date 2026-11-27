import{t as e}from"./chunk-JWPE2WC7.Otrbkxt-.js";import{In as t,Kn as n,Nn as r,Nt as i,Pn as a,Sn as o,Yn as s,_n as c,bn as l,mn as u,vt as d,wn as f,xn as p,yn as m}from"./theme.DFPze7FE.js";import{n as h}from"./mermaid-parser.core.DoQxVA4O.js";var g=m.packet,_=class{constructor(){this.packet=[],this.setAccTitle=a,this.getAccTitle=p,this.setDiagramTitle=t,this.getDiagramTitle=f,this.getAccDescription=l,this.setAccDescription=r}static{s(this,`PacketDB`)}getConfig(){let e=d({...g,...o().packet});return e.showBits&&(e.paddingY+=10),e}getPacket(){return this.packet}pushWord(e){e.length>0&&this.packet.push(e)}clear(){u(),this.packet=[]}},v=1e4,y=s((t,r)=>{e(t,r);let i=-1,a=[],o=1,{bitsPerRow:s}=r.getConfig();for(let{start:e,end:c,bits:l,label:u}of t.blocks){if(e!==void 0&&c!==void 0&&c<e)throw Error(`Packet block ${e} - ${c} is invalid. End must be greater than start.`);if(e??=i+1,e!==i+1)throw Error(`Packet block ${e} - ${c??e} is not contiguous. It should start from ${i+1}.`);if(l===0)throw Error(`Packet block ${e} is invalid. Cannot have a zero bit field.`);for(c??=e+(l??1)-1,l??=c-e+1,i=c,n.debug(`Packet block ${e} - ${i} with label ${u}`);a.length<=s+1&&r.getPacket().length<v;){let[t,n]=b({start:e,end:c,bits:l,label:u},o,s);if(a.push(t),t.end+1===o*s&&(r.pushWord(a),a=[],o++),!n)break;({start:e,end:c,bits:l,label:u}=n)}}r.pushWord(a)},`populate`),b=s((e,t,n)=>{if(e.start===void 0)throw Error(`start should have been set during first phase`);if(e.end===void 0)throw Error(`end should have been set during first phase`);if(e.start>e.end)throw Error(`Block start ${e.start} is greater than block end ${e.end}.`);if(e.end+1<=t*n)return[e,void 0];let r=t*n-1,i=t*n;return[{start:e.start,end:r,label:e.label,bits:r-e.start},{start:i,end:e.end,label:e.label,bits:e.end-i}]},`getNextFittingBlock`),x={parser:{yy:void 0},parse:s(async e=>{let t=await h(`packet`,e),r=x.parser?.yy;if(!(r instanceof _))throw Error(`parser.parser?.yy was not a PacketDB. This is due to a bug within Mermaid, please report this issue at https://github.com/mermaid-js/mermaid/issues.`);n.debug(t),y(t,r)},`parse`)},S=s((e,t,n,r)=>{let a=r.db,o=a.getConfig(),{rowHeight:s,paddingY:l,bitWidth:u,bitsPerRow:d}=o,f=a.getPacket(),p=a.getDiagramTitle(),m=s+l,h=m*(f.length+1)-(p?0:s),g=u*d+2,_=i(t);_.attr(`viewBox`,`0 0 ${g} ${h}`),c(_,h,g,o.useMaxWidth);for(let[e,t]of f.entries())C(_,t,e,o);_.append(`text`).text(p).attr(`x`,g/2).attr(`y`,h-m/2).attr(`dominant-baseline`,`middle`).attr(`text-anchor`,`middle`).attr(`class`,`packetTitle`)},`draw`),C=s((e,t,n,{rowHeight:r,paddingX:i,paddingY:a,bitWidth:o,bitsPerRow:s,showBits:c})=>{let l=e.append(`g`),u=n*(r+a)+a;for(let e of t){let t=e.start%s*o+1,n=(e.end-e.start+1)*o-i;if(l.append(`rect`).attr(`x`,t).attr(`y`,u).attr(`width`,n).attr(`height`,r).attr(`class`,`packetBlock`),l.append(`text`).attr(`x`,t+n/2).attr(`y`,u+r/2).attr(`class`,`packetLabel`).attr(`dominant-baseline`,`middle`).attr(`text-anchor`,`middle`).text(e.label),!c)continue;let a=e.end===e.start,d=u-2;l.append(`text`).attr(`x`,t+(a?n/2:0)).attr(`y`,d).attr(`class`,`packetByte start`).attr(`dominant-baseline`,`auto`).attr(`text-anchor`,a?`middle`:`start`).text(e.start),a||l.append(`text`).attr(`x`,t+n).attr(`y`,d).attr(`class`,`packetByte end`).attr(`dominant-baseline`,`auto`).attr(`text-anchor`,`end`).text(e.end)}},`drawWord`),w={draw:S},T={byteFontSize:`10px`,startByteColor:`black`,endByteColor:`black`,labelColor:`black`,labelFontSize:`12px`,titleColor:`black`,titleFontSize:`14px`,blockStrokeColor:`black`,blockStrokeWidth:`1`,blockFillColor:`#efefef`},E={parser:x,get db(){return new _},renderer:w,styles:s(({packet:e}={})=>{let t=d(T,e);return`
	.packetByte {
		font-size: ${t.byteFontSize};
	}
	.packetByte.start {
		fill: ${t.startByteColor};
	}
	.packetByte.end {
		fill: ${t.endByteColor};
	}
	.packetLabel {
		fill: ${t.labelColor};
		font-size: ${t.labelFontSize};
	}
	.packetTitle {
		fill: ${t.titleColor};
		font-size: ${t.titleFontSize};
	}
	.packetBlock {
		stroke: ${t.blockStrokeColor};
		stroke-width: ${t.blockStrokeWidth};
		fill: ${t.blockFillColor};
	}
	`},`styles`)};export{E as diagram};