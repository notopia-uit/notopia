import{Fn as e,In as t,It as n,P as r,Pt as i,Rn as a,Rt as o,Sn as s,Tn as c,Xn as l,Yn as u,bn as d,hn as f,rn as p,vn as m,wn as h,xn as g,z as _}from"./theme.DHeniVf8.js";import{t as v}from"./mermaid-parser.core.z_957cbL.js";import{t as y}from"./chunk-4BX2VUAB.BFlTnfga.js";var b=d.pie,x={sections:new Map,showData:!1,config:b},S=x.sections,C=x.showData,w=structuredClone(b),T={getConfig:u(()=>structuredClone(w),`getConfig`),clear:u(()=>{S=new Map,C=x.showData,f()},`clear`),setDiagramTitle:a,getDiagramTitle:c,setAccTitle:t,getAccTitle:s,setAccDescription:e,getAccDescription:g,addSection:u(({label:e,value:t})=>{if(t<0)throw Error(`"${e}" has invalid value: ${t}. Negative values are not allowed in pie charts. All slice values must be >= 0.`);S.has(e)||(S.set(e,t),l.debug(`added new section: ${e}, with value: ${t}`))},`addSection`),getSections:u(()=>S,`getSections`),setShowData:u(e=>{C=e},`setShowData`),getShowData:u(()=>C,`getShowData`)},E=u((e,t)=>{y(e,t),t.setShowData(e.showData),e.sections.map(t.addSection)},`populateDb`),D={parse:u(async e=>{let t=await v(`pie`,e);l.debug(t),E(t,T)},`parse`)},O=u(e=>`
  .pieCircle{
    stroke: ${e.pieStrokeColor};
    stroke-width : ${e.pieStrokeWidth};
    opacity : ${e.pieOpacity};
  }
  .pieOuterCircle{
    stroke: ${e.pieOuterStrokeColor};
    stroke-width: ${e.pieOuterStrokeWidth};
    fill: none;
  }
  .pieTitleText {
    text-anchor: middle;
    font-size: ${e.pieTitleTextSize};
    fill: ${e.pieTitleTextColor};
    font-family: ${e.fontFamily};
  }
  .slice {
    font-family: ${e.fontFamily};
    fill: ${e.pieSectionTextColor};
    font-size:${e.pieSectionTextSize};
    // fill: white;
  }
  .legend text {
    fill: ${e.pieLegendTextColor};
    font-family: ${e.fontFamily};
    font-size: ${e.pieLegendTextSize};
  }
`,`getStyles`),k=u(e=>{let t=[...e.values()].reduce((e,t)=>e+t,0),r=[...e.entries()].map(([e,t])=>({label:e,value:t})).filter(e=>e.value/t*100>=1).sort((e,t)=>t.value-e.value);return n().value(e=>e.value)(r)},`createPieArcs`),A={parser:D,db:T,renderer:{draw:u((e,t,n,a)=>{l.debug(`rendering pie chart
`+e);let s=a.db,c=h(),u=r(s.getConfig(),c.pie),d=i(t),f=d.append(`g`);f.attr(`transform`,`translate(225,225)`);let{themeVariables:g}=c,[v]=_(g.pieOuterStrokeWidth);v??=2;let y=u.textPosition,b=o().innerRadius(0).outerRadius(185),x=o().innerRadius(185*y).outerRadius(185*y);f.append(`circle`).attr(`cx`,0).attr(`cy`,0).attr(`r`,185+v/2).attr(`class`,`pieOuterCircle`);let S=s.getSections(),C=k(S),w=[g.pie1,g.pie2,g.pie3,g.pie4,g.pie5,g.pie6,g.pie7,g.pie8,g.pie9,g.pie10,g.pie11,g.pie12],T=0;S.forEach(e=>{T+=e});let E=C.filter(e=>(e.data.value/T*100).toFixed(0)!==`0`),D=p(w);f.selectAll(`mySlices`).data(E).enter().append(`path`).attr(`d`,b).attr(`fill`,e=>D(e.data.label)).attr(`class`,`pieCircle`),f.selectAll(`mySlices`).data(E).enter().append(`text`).text(e=>(e.data.value/T*100).toFixed(0)+`%`).attr(`transform`,e=>`translate(`+x.centroid(e)+`)`).style(`text-anchor`,`middle`).attr(`class`,`slice`),f.append(`text`).text(s.getDiagramTitle()).attr(`x`,0).attr(`y`,-200).attr(`class`,`pieTitleText`);let O=[...S.entries()].map(([e,t])=>({label:e,value:t})),A=f.selectAll(`.legend`).data(O).enter().append(`g`).attr(`class`,`legend`).attr(`transform`,(e,t)=>{let n=22*O.length/2;return`translate(216,`+(t*22-n)+`)`});A.append(`rect`).attr(`width`,18).attr(`height`,18).style(`fill`,e=>D(e.label)).style(`stroke`,e=>D(e.label)),A.append(`text`).attr(`x`,22).attr(`y`,14).text(e=>s.getShowData()?`${e.label} [${e.value}]`:e.label);let j=512+Math.max(...A.selectAll(`text`).nodes().map(e=>e?.getBoundingClientRect().width??0));d.attr(`viewBox`,`0 0 ${j} 450`),m(d,450,j,u.useMaxWidth)},`draw`)},styles:O};export{A as diagram};