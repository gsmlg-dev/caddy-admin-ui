(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[83],{9837:function(e,t,n){"use strict";n.d(t,{Z:function(){return g}});var r=n(7462),o=n(3366),a=n(7294),i=n(6010),s=n(4780),l=n(1719),c=n(8884),f=n(1401),u=n(1588),d=n(4867);function h(e){return(0,d.Z)("MuiCard",e)}(0,u.Z)("MuiCard",["root"]);var p=n(5893);let m=["className","raised"],Z=e=>{let{classes:t}=e;return(0,s.Z)({root:["root"]},h,t)},x=(0,l.ZP)(f.Z,{name:"MuiCard",slot:"Root",overridesResolver:(e,t)=>t.root})(()=>({overflow:"hidden"})),v=a.forwardRef(function(e,t){let n=(0,c.Z)({props:e,name:"MuiCard"}),{className:a,raised:s=!1}=n,l=(0,o.Z)(n,m),f=(0,r.Z)({},n,{raised:s}),u=Z(f);return(0,p.jsx)(x,(0,r.Z)({className:(0,i.Z)(u.root,a),elevation:s?8:void 0,ref:t,ownerState:f},l))});var g=v},1359:function(e,t,n){"use strict";n.d(t,{Z:function(){return v}});var r=n(7462),o=n(3366),a=n(7294),i=n(6010),s=n(4780),l=n(1719),c=n(8884),f=n(1588),u=n(4867);function d(e){return(0,u.Z)("MuiCardContent",e)}(0,f.Z)("MuiCardContent",["root"]);var h=n(5893);let p=["className","component"],m=e=>{let{classes:t}=e;return(0,s.Z)({root:["root"]},d,t)},Z=(0,l.ZP)("div",{name:"MuiCardContent",slot:"Root",overridesResolver:(e,t)=>t.root})(()=>({padding:16,"&:last-child":{paddingBottom:24}})),x=a.forwardRef(function(e,t){let n=(0,c.Z)({props:e,name:"MuiCardContent"}),{className:a,component:s="div"}=n,l=(0,o.Z)(n,p),f=(0,r.Z)({},n,{component:s}),u=m(f);return(0,h.jsx)(Z,(0,r.Z)({as:s,className:(0,i.Z)(u.root,a),ownerState:f,ref:t},l))});var v=x},5048:function(e,t,n){(window.__NEXT_P=window.__NEXT_P||[]).push(["/upstream",function(){return n(4264)}])},4264:function(e,t,n){"use strict";n.r(t),n.d(t,{default:function(){return h}});var r=n(5893),o=n(7294),a=n(6336),i=n(9630),s=n(1953),l=n(9465),c=n(1906),f=n(9837),u=n(1359);function d(e){let{address:t,healthy:n,num_requests:o,fails:a}=e;return(0,r.jsx)(f.Z,{sx:{minWidth:275},children:(0,r.jsxs)(u.Z,{children:[(0,r.jsxs)(i.Z,{sx:{fontSize:14},color:"text.secondary",gutterBottom:!0,children:["Healthy: ",n?"\uD83D\uDC4D":"\uD83D\uDC4E"]}),(0,r.jsxs)(i.Z,{variant:"h5",component:"div",children:["Address: ",t]}),(0,r.jsxs)(i.Z,{variant:"body2",children:["Requests: ".concat(o),(0,r.jsx)("br",{}),"Fails: ".concat(a)]})]})})}function h(){let[e,t]=o.useState([]);return o.useEffect(()=>{let e=async()=>{let e=await fetch("/reverse_proxy/upstreams"),n=await e.json();t(n)};e()},[]),(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(l.Z,{}),(0,r.jsx)(a.Z,{maxWidth:"lg",children:(0,r.jsxs)(s.Z,{sx:{my:4},children:[(0,r.jsx)(i.Z,{variant:"h4",component:"h1",gutterBottom:!0,children:"Caddy Server Upstream"}),(0,r.jsx)(i.Z,{variant:"p",component:"p",gutterBottom:!0,children:e.map(e=>(0,r.jsx)(d,{...e}))}),(0,r.jsx)(c.Z,{})]})})]})}},9465:function(e,t,n){"use strict";n.d(t,{Z:function(){return T}});var r=n(7297),o=n(5944),a=n(917),i=n(7294),s=n(5050),l=n(1953),c=n(784),f=n(562),u=n(9630),d=n(9934),h=n(326),p=n(6336),m=n(5084),Z=n(7056),x=n(3960),v=n(5893),g=n(5697),y=n.n(g),j=n(6010),b=n(1163),C=n(1664),w=n.n(C),k=n(8346),N=n(1719);let _=(0,N.ZP)("a")({}),D=i.forwardRef(function(e,t){let{to:n,linkAs:r,replace:o,scroll:a,shallow:i,prefetch:s,locale:l,...c}=e;return(0,v.jsx)(w(),{href:n,prefetch:s,as:r,replace:o,scroll:a,shallow:i,passHref:!0,locale:l,children:(0,v.jsx)(_,{ref:t,...c})})});D.propTypes={href:y().any,linkAs:y().oneOfType([y().object,y().string]),locale:y().string,passHref:y().bool,prefetch:y().bool,replace:y().bool,scroll:y().bool,shallow:y().bool,to:y().oneOfType([y().object,y().string]).isRequired};let M=i.forwardRef(function(e,t){let{activeClassName:n="active",as:r,className:o,href:a,linkAs:i,locale:s,noLinkStyle:l,prefetch:c,replace:f,role:u,scroll:d,shallow:h,...p}=e,m=(0,b.useRouter)(),Z="string"==typeof a?a:a.pathname,x=(0,j.Z)(o,{[n]:m.pathname===Z&&n}),g="string"==typeof a&&(0===a.indexOf("http")||0===a.indexOf("mailto:"));if(g)return l?(0,v.jsx)(_,{className:x,href:a,ref:t,...p}):(0,v.jsx)(k.Z,{className:x,href:a,ref:t,...p});let y={to:a,linkAs:i||r,replace:f,scroll:d,shallow:h,prefetch:c,locale:s};return l?(0,v.jsx)(D,{className:x,ref:t,...y,...p}):(0,v.jsx)(k.Z,{component:D,className:x,ref:t,...y,...p})});function R(){let e=(0,r.Z)(["\n                      color: white;\n                    "]);return R=function(){return e},e}function O(){let e=(0,r.Z)(["\n                    color: white;\n                  "]);return O=function(){return e},e}M.propTypes={activeClassName:y().string,as:y().oneOfType([y().object,y().string]),className:y().string,href:y().any,linkAs:y().oneOfType([y().object,y().string]),locale:y().string,noLinkStyle:y().bool,prefetch:y().bool,replace:y().bool,role:y().string,scroll:y().bool,shallow:y().bool};let S=[{name:"Upstream",href:"/upstream"},{name:"PKI",href:"/pki-view"},{name:"Load",href:"/setup"},{name:"Adapt",href:"/convert-config"},{name:"Metrics",href:"/monitor"}],E=()=>{let[e,t]=i.useState(null),n=e=>{t(e.currentTarget)},r=()=>{t(null)};return(0,o.tZ)(s.Z,{position:"static",children:(0,o.tZ)(p.Z,{maxWidth:"lg",children:(0,o.BX)(c.Z,{disableGutters:!0,children:[(0,o.tZ)(x.Z,{sx:{display:{xs:"none",md:"flex"},mr:1}}),(0,o.tZ)(u.Z,{variant:"h6",noWrap:!0,component:"a",href:"/",sx:{mr:2,display:{xs:"none",md:"flex"},fontFamily:"monospace",fontWeight:700,letterSpacing:".3rem",color:"inherit",textDecoration:"none"},children:"Caddy Config"}),(0,o.BX)(l.Z,{sx:{flexGrow:1,display:{xs:"flex",md:"none"}},children:[(0,o.tZ)(f.Z,{size:"large","aria-label":"account of current user","aria-controls":"menu-appbar","aria-haspopup":"true",onClick:n,color:"inherit",children:(0,o.tZ)(h.Z,{})}),(0,o.tZ)(d.Z,{id:"menu-appbar",anchorEl:e,anchorOrigin:{vertical:"bottom",horizontal:"left"},keepMounted:!0,transformOrigin:{vertical:"top",horizontal:"left"},open:Boolean(e),onClose:r,sx:{display:{xs:"block",md:"none"}},children:S.map(e=>(0,o.tZ)(Z.Z,{onClick:r,children:(0,o.tZ)(M,{href:e.href,css:(0,a.iv)(R()),children:(0,o.tZ)(u.Z,{textAlign:"center",children:e.name})})},e.name))})]}),(0,o.tZ)(x.Z,{sx:{display:{xs:"flex",md:"none"},mr:1}}),(0,o.tZ)(u.Z,{variant:"h5",noWrap:!0,component:"a",href:"",sx:{mr:2,display:{xs:"flex",md:"none"},flexGrow:1,fontFamily:"monospace",fontWeight:700,letterSpacing:".3rem",color:"inherit",textDecoration:"none"},children:"Caddy Config"}),(0,o.tZ)(l.Z,{sx:{flexGrow:1,display:{xs:"none",md:"flex"}},children:S.map(e=>(0,o.tZ)(m.Z,{onClick:r,sx:{my:2,color:"white",display:"block"},children:(0,o.tZ)(M,{href:e.href,css:(0,a.iv)(O()),children:e.name})},e.name))})]})})})};var T=E},1906:function(e,t,n){"use strict";n.d(t,{Z:function(){return i}});var r=n(5893);n(7294);var o=n(9630),a=n(8346);function i(){return(0,r.jsxs)(o.Z,{variant:"body2",color:"text.secondary",align:"center",children:["Copyright \xa9 ",(0,r.jsx)(a.Z,{color:"inherit",href:"https://github.com/gsmlg-dev/",children:"GSMLG-DEV"})," ",new Date().getFullYear(),"."]})}}},function(e){e.O(0,[514,774,888,179],function(){return e(e.s=5048)}),_N_E=e.O()}]);