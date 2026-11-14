import{t as e}from"./chunk-JWPE2WC7.DQ-6lTZk.js";import{Bn as t,Dn as n,Ln as r,Tn as i,an as a,cn as o,dn as s,fn as c,hn as l,lt as u,pn as d,un as f,wn as p,wt as m}from"./theme.CuQJBZVz.js";import{n as h}from"./mermaid-parser.core.BhAbpN2P.js";var g=f.packet,_=class{constructor(){this.packet=[],this.setAccTitle=i,this.getAccTitle=c,this.setDiagramTitle=n,this.getDiagramTitle=l,this.getAccDescription=s,this.setAccDescription=p}static{t(this,`PacketDB`)}getConfig(){let e=u({...g,...d().packet});return e.showBits&&(e.paddingY+=10),e}getPacket(){return this.packet}pushWord(e){e.length>0&&this.packet.push(e)}clear(){a(),this.packet=[]}},v=1e4,y=t((t,n)=>{e(t,n);let i=-1,a=[],o=1,{bitsPerRow:s}=n.getConfig();for(let{start:e,end:c,bits:l,label:u}of t.blocks){if(e!==void 0&&c!==void 0&&c<e)throw Error(`Packet block ${e} - ${c} is invalid. End must be greater than start.`);if(e??=i+1,e!==i+1)throw Error(`Packet block ${e} - ${c??e} is not contiguous. It should start from ${i+1}.`);if(l===0)throw Error(`Packet block ${e} is invalid. Cannot have a zero bit field.`);for(c??=e+(l??1)-1,l??=c-e+1,i=c,r.debug(`Packet block ${e} - ${i} with label ${u}`);a.length<=s+1&&n.getPacket().length<v;){let[t,r]=b({start:e,end:c,bits:l,label:u},o,s);if(a.push(t),t.end+1===o*s&&(n.pushWord(a),a=[],o++),!r)break;({start:e,end:c,bits:l,label:u}=r)}}n.pushWord(a)},`populate`),b=t((e,t,n)=>{if(e.start===void 0)throw Error(`start should have been set during first phase`);if(e.end===void 0)throw Error(`end should have been set during first phase`);if(e.start>e.end)throw Error(`Block start ${e.start} is greater than block end ${e.end}.`);if(e.end+1<=t*n)return[e,void 0];let r=t*n-1,i=t*n;return[{start:e.start,end:r,label:e.label,bits:r-e.start},{start:i,end:e.end,label:e.label,bits:e.end-i}]},`getNextFittingBlock`),x={parser:{yy:void 0},parse:t(async e=>{let t=await h(`packet`,e),n=x.parser?.yy;if(!(n instanceof _))throw Error(`parser.parser?.yy was not a PacketDB. This is due to a bug within Mermaid, please report this issue at https://github.com/mermaid-js/mermaid/issues.`);r.debug(t),y(t,n)},`parse`)},S=t((e,t,n,r)=>{let i=r.db,a=i.getConfig(),{rowHeight:s,paddingY:c,bitWidth:l,bitsPerRow:u}=a,d=i.getPacket(),f=i.getDiagramTitle(),p=s+c,h=p*(d.length+1)-(f?0:s),g=l*u+2,_=m(t);_.attr(`viewBox`,`0 0 ${g} ${h}`),o(_,h,g,a.useMaxWidth);for(let[e,t]of d.entries())C(_,t,e,a);_.append(`text`).text(f).attr(`x`,g/2).attr(`y`,h-p/2).attr(`dominant-baseline`,`middle`).attr(`text-anchor`,`middle`).attr(`class`,`packetTitle`)},`draw`),C=t((e,t,n,{rowHeight:r,paddingX:i,paddingY:a,bitWidth:o,bitsPerRow:s,showBits:c})=>{let l=e.append(`g`),u=n*(r+a)+a;for(let e of t){let t=e.start%s*o+1,n=(e.end-e.start+1)*o-i;if(l.append(`rect`).attr(`x`,t).attr(`y`,u).attr(`width`,n).attr(`height`,r).attr(`class`,`packetBlock`),l.append(`text`).attr(`x`,t+n/2).attr(`y`,u+r/2).attr(`class`,`packetLabel`).attr(`dominant-baseline`,`middle`).attr(`text-anchor`,`middle`).text(e.label),!c)continue;let a=e.end===e.start,d=u-2;l.append(`text`).attr(`x`,t+(a?n/2:0)).attr(`y`,d).attr(`class`,`packetByte start`).attr(`dominant-baseline`,`auto`).attr(`text-anchor`,a?`middle`:`start`).text(e.start),a||l.append(`text`).attr(`x`,t+n).attr(`y`,d).attr(`class`,`packetByte end`).attr(`dominant-baseline`,`auto`).attr(`text-anchor`,`end`).text(e.end)}},`drawWord`),w={draw:S},T={byteFontSize:`10px`,startByteColor:`black`,endByteColor:`black`,labelColor:`black`,labelFontSize:`12px`,titleColor:`black`,titleFontSize:`14px`,blockStrokeColor:`black`,blockStrokeWidth:`1`,blockFillColor:`#efefef`},E={parser:x,get db(){return new _},renderer:w,styles:t(({packet:e}={})=>{let t=u(T,e);return`
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