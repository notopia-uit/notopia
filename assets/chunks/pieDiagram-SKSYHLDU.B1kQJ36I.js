import{$t as e,B as t,Cn as n,En as r,F as i,Ft as a,In as o,Ln as s,Sn as c,Tn as l,Xn as u,Xt as d,Zn as f,gn as p,in as m,xn as h,yn as g,zn as _}from"./theme.DJ99bU_F.js";import{t as v}from"./mermaid-parser.core.DdlAO6AV.js";import{t as y}from"./chunk-4BX2VUAB.CFsHq6B0.js";var b=h.pie,x={sections:new Map,showData:!1,config:b},S=x.sections,C=x.showData,w=structuredClone(b),T={getConfig:u(()=>structuredClone(w),`getConfig`),clear:u(()=>{S=new Map,C=x.showData,p()},`clear`),setDiagramTitle:_,getDiagramTitle:r,setAccTitle:s,getAccTitle:n,setAccDescription:o,getAccDescription:c,addSection:u(({label:e,value:t})=>{if(t<0)throw Error(`"${e}" has invalid value: ${t}. Negative values are not allowed in pie charts. All slice values must be >= 0.`);S.has(e)||(S.set(e,t),f.debug(`added new section: ${e}, with value: ${t}`))},`addSection`),getSections:u(()=>S,`getSections`),setShowData:u(e=>{C=e},`setShowData`),getShowData:u(()=>C,`getShowData`)},E=u((e,t)=>{y(e,t),t.setShowData(e.showData),e.sections.map(t.addSection)},`populateDb`),D={parse:u(async e=>{let t=await v(`pie`,e);f.debug(t),E(t,T)},`parse`)},O=u(e=>`
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
`,`getStyles`),k=u(t=>{let n=[...t.values()].reduce((e,t)=>e+t,0),r=[...t.entries()].map(([e,t])=>({label:e,value:t})).filter(e=>e.value/n*100>=1).sort((e,t)=>t.value-e.value);return e().value(e=>e.value)(r)},`createPieArcs`),A={parser:D,db:T,renderer:{draw:u((e,n,r,o)=>{f.debug(`rendering pie chart
`+e);let s=o.db,c=l(),u=i(s.getConfig(),c.pie),p=a(n),h=p.append(`g`);h.attr(`transform`,`translate(225,225)`);let{themeVariables:_}=c,[v]=t(_.pieOuterStrokeWidth);v??=2;let y=u.textPosition,b=d().innerRadius(0).outerRadius(185),x=d().innerRadius(185*y).outerRadius(185*y);h.append(`circle`).attr(`cx`,0).attr(`cy`,0).attr(`r`,185+v/2).attr(`class`,`pieOuterCircle`);let S=s.getSections(),C=k(S),w=[_.pie1,_.pie2,_.pie3,_.pie4,_.pie5,_.pie6,_.pie7,_.pie8,_.pie9,_.pie10,_.pie11,_.pie12],T=0;S.forEach(e=>{T+=e});let E=C.filter(e=>(e.data.value/T*100).toFixed(0)!==`0`),D=m(w);h.selectAll(`mySlices`).data(E).enter().append(`path`).attr(`d`,b).attr(`fill`,e=>D(e.data.label)).attr(`class`,`pieCircle`),h.selectAll(`mySlices`).data(E).enter().append(`text`).text(e=>(e.data.value/T*100).toFixed(0)+`%`).attr(`transform`,e=>`translate(`+x.centroid(e)+`)`).style(`text-anchor`,`middle`).attr(`class`,`slice`),h.append(`text`).text(s.getDiagramTitle()).attr(`x`,0).attr(`y`,-200).attr(`class`,`pieTitleText`);let O=[...S.entries()].map(([e,t])=>({label:e,value:t})),A=h.selectAll(`.legend`).data(O).enter().append(`g`).attr(`class`,`legend`).attr(`transform`,(e,t)=>{let n=22*O.length/2;return`translate(216,`+(t*22-n)+`)`});A.append(`rect`).attr(`width`,18).attr(`height`,18).style(`fill`,e=>D(e.label)).style(`stroke`,e=>D(e.label)),A.append(`text`).attr(`x`,22).attr(`y`,14).text(e=>s.getShowData()?`${e.label} [${e.value}]`:e.label);let j=512+Math.max(...A.selectAll(`text`).nodes().map(e=>e?.getBoundingClientRect().width??0));p.attr(`viewBox`,`0 0 ${j} 450`),g(p,450,j,u.useMaxWidth)},`draw`)},styles:O};export{A as diagram};