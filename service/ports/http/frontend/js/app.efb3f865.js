(function(){"use strict";var e={9305:function(e,t,r){var n=r(9242),o=r(3396),i=r(2268);function s(e,t,r,n,s,c){const u=(0,o.up)("router-view");return(0,o.wg)(),(0,o.iD)(o.HY,null,[(0,o._)("h1",null," Nos crossposting service "+(0,i.zw)(e.user),1),(0,o.Wm)(u)],64)}var c,u=r(7327),l=r(6520),a=r(7139),p=r(4161);(function(e){e["SetUser"]="setUser"})(c||(c={}));var f=(0,a.MT)({state:{user:void 0},getters:{},mutations:{[c.SetUser](e,t){e.user=t}},actions:{},modules:{}});class d{constructor(e){(0,u.Z)(this,"store",void 0),(0,u.Z)(this,"axios",p.Z.create()),this.store=e}currentUser(){const e="/api/current-user";return this.axios.get(e)}publicKeys(){const e="/api/public-keys";return this.axios.get(e)}addPublicKey(e){const t="/api/public-keys";return this.axios.post(t,e)}refreshCurrentUser(){return new Promise(((e,t)=>{this.currentUser().then((t=>{this.store.commit(c.SetUser,t.data.user),e(t.data)}),(e=>{t(e)}))}))}}var v=function(e,t,r,n){var o,i=arguments.length,s=i<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"===typeof Reflect&&"function"===typeof Reflect.decorate)s=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(o=e[c])&&(s=(i<3?o(s):i>3?o(t,r,s):o(t,r))||s);return i>3&&s&&Object.defineProperty(t,r,s),s};let h=class extends l.w3{constructor(...e){super(...e),(0,u.Z)(this,"apiService",new d((0,a.oR)()))}created(){this.loadCurrentUser()}loadCurrentUser(){this.apiService.refreshCurrentUser()}};h=v([(0,l.Ei)({})],h);var g=h,y=r(89);const b=(0,y.Z)(g,[["render",s]]);var w=b,k=r(2483);const O={class:"home"},D={key:0},j={key:1},m={key:2},P={key:0},x={key:1},K={key:0},R={key:1};function Z(e,t,r,s,c,u){const l=(0,o.up)("Explanation"),a=(0,o.up)("LogInWithTwitterButton"),p=(0,o.up)("CurrentUser");return(0,o.wg)(),(0,o.iD)("div",O,[e.loading?((0,o.wg)(),(0,o.iD)("div",D," Loading... ")):(0,o.kq)("",!0),e.loading||e.user?(0,o.kq)("",!0):((0,o.wg)(),(0,o.iD)("div",j,[(0,o.Wm)(l),(0,o.Wm)(a)])),!e.loading&&e.user?((0,o.wg)(),(0,o.iD)("div",m,[(0,o.Wm)(p,{user:e.user},null,8,["user"]),e.publicKeys?(0,o.kq)("",!0):((0,o.wg)(),(0,o.iD)("div",P," Loading public keys... ")),e.publicKeys?((0,o.wg)(),(0,o.iD)("div",x,[e.publicKeys.publicKeys?.length>0?((0,o.wg)(),(0,o.iD)("ul",K,[((0,o.wg)(!0),(0,o.iD)(o.HY,null,(0,o.Ko)(e.publicKeys.publicKeys,(e=>((0,o.wg)(),(0,o.iD)("li",{key:e.npub},(0,i.zw)(e.npub),1)))),128))])):(0,o.kq)("",!0),0==e.publicKeys.publicKeys?.length?((0,o.wg)(),(0,o.iD)("p",R," You haven't added any public keys yet. ")):(0,o.kq)("",!0)])):(0,o.kq)("",!0)])):(0,o.kq)("",!0),(0,o.wy)((0,o._)("input",{placeholder:"npub...","onUpdate:modelValue":t[0]||(t[0]=t=>e.npub=t)},null,512),[[n.nr,e.npub]]),(0,o._)("button",{onClick:t[1]||(t[1]=(...t)=>e.addPublicKey&&e.addPublicKey(...t))},"Link public key")])}const U={class:"explanation"},_=(0,o._)("p",null," Explanation. ",-1),C=[_];function E(e,t,r,n,i,s){return(0,o.wg)(),(0,o.iD)("div",U,C)}var S=function(e,t,r,n){var o,i=arguments.length,s=i<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"===typeof Reflect&&"function"===typeof Reflect.decorate)s=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(o=e[c])&&(s=(i<3?o(s):i>3?o(t,r,s):o(t,r))||s);return i>3&&s&&Object.defineProperty(t,r,s),s};let q=class extends l.w3{};q=S([(0,l.Ei)({})],q);var L=q;const T=(0,y.Z)(L,[["render",E]]);var W=T;const I={class:"log-in-with-twitter-button"},Y=(0,o._)("a",{href:"/login"},"Log in with Twitter.",-1),z=[Y];function B(e,t,r,n,i,s){return(0,o.wg)(),(0,o.iD)("div",I,z)}var H=function(e,t,r,n){var o,i=arguments.length,s=i<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"===typeof Reflect&&"function"===typeof Reflect.decorate)s=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(o=e[c])&&(s=(i<3?o(s):i>3?o(t,r,s):o(t,r))||s);return i>3&&s&&Object.defineProperty(t,r,s),s};let M=class extends l.w3{};M=H([(0,l.Ei)({})],M);var F=M;const N=(0,y.Z)(F,[["render",B]]);var V=N;const A={class:"current-user"};function G(e,t,r,n,s,c){return(0,o.wg)(),(0,o.iD)("div",A," You are logged in as "+(0,i.zw)(e.user.accountID)+". ",1)}class J{constructor(){(0,u.Z)(this,"accountID",void 0),(0,u.Z)(this,"twitterID",void 0)}}var Q=function(e,t,r,n){var o,i=arguments.length,s=i<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"===typeof Reflect&&"function"===typeof Reflect.decorate)s=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(o=e[c])&&(s=(i<3?o(s):i>3?o(t,r,s):o(t,r))||s);return i>3&&s&&Object.defineProperty(t,r,s),s};let X=class extends l.w3{constructor(...e){super(...e),(0,u.Z)(this,"user",void 0)}};X=Q([(0,l.Ei)({props:{user:J}})],X);var $=X;const ee=(0,y.Z)($,[["render",G]]);var te=ee;class re{constructor(e){(0,u.Z)(this,"npub",void 0),this.npub=e}}var ne=function(e,t,r,n){var o,i=arguments.length,s=i<3?t:null===n?n=Object.getOwnPropertyDescriptor(t,r):n;if("object"===typeof Reflect&&"function"===typeof Reflect.decorate)s=Reflect.decorate(e,t,r,n);else for(var c=e.length-1;c>=0;c--)(o=e[c])&&(s=(i<3?o(s):i>3?o(t,r,s):o(t,r))||s);return i>3&&s&&Object.defineProperty(t,r,s),s};let oe=class extends l.w3{constructor(...e){super(...e),(0,u.Z)(this,"apiService",new d((0,a.oR)())),(0,u.Z)(this,"store",(0,a.oR)()),(0,u.Z)(this,"publicKeys",null),(0,u.Z)(this,"npub","")}get loading(){return void 0===this.store.state.user}get user(){return this.store.state.user}created(){this.apiService.publicKeys().then((e=>{this.publicKeys=e.data}))}addPublicKey(){this.apiService.addPublicKey(new re(this.npub)).then((e=>{console.log("added")})).catch((e=>{console.log("error")}))}};oe=ne([(0,l.Ei)({components:{CurrentUser:te,LogInWithTwitterButton:V,Explanation:W}})],oe);var ie=oe;const se=(0,y.Z)(ie,[["render",Z]]);var ce=se;const ue=[{path:"/",name:"home",component:ce}],le=(0,k.p7)({history:(0,k.PO)("/"),routes:ue});var ae=le;(0,n.ri)(w).use(f).use(ae).mount("#app")}},t={};function r(n){var o=t[n];if(void 0!==o)return o.exports;var i=t[n]={exports:{}};return e[n].call(i.exports,i,i.exports,r),i.exports}r.m=e,function(){var e=[];r.O=function(t,n,o,i){if(!n){var s=1/0;for(a=0;a<e.length;a++){n=e[a][0],o=e[a][1],i=e[a][2];for(var c=!0,u=0;u<n.length;u++)(!1&i||s>=i)&&Object.keys(r.O).every((function(e){return r.O[e](n[u])}))?n.splice(u--,1):(c=!1,i<s&&(s=i));if(c){e.splice(a--,1);var l=o();void 0!==l&&(t=l)}}return t}i=i||0;for(var a=e.length;a>0&&e[a-1][2]>i;a--)e[a]=e[a-1];e[a]=[n,o,i]}}(),function(){r.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return r.d(t,{a:t}),t}}(),function(){r.d=function(e,t){for(var n in t)r.o(t,n)&&!r.o(e,n)&&Object.defineProperty(e,n,{enumerable:!0,get:t[n]})}}(),function(){r.g=function(){if("object"===typeof globalThis)return globalThis;try{return this||new Function("return this")()}catch(e){if("object"===typeof window)return window}}()}(),function(){r.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)}}(),function(){var e={143:0};r.O.j=function(t){return 0===e[t]};var t=function(t,n){var o,i,s=n[0],c=n[1],u=n[2],l=0;if(s.some((function(t){return 0!==e[t]}))){for(o in c)r.o(c,o)&&(r.m[o]=c[o]);if(u)var a=u(r)}for(t&&t(n);l<s.length;l++)i=s[l],r.o(e,i)&&e[i]&&e[i][0](),e[i]=0;return r.O(a)},n=self["webpackChunknos_crossposting_service_frontend"]=self["webpackChunknos_crossposting_service_frontend"]||[];n.forEach(t.bind(null,0)),n.push=t.bind(null,n.push.bind(n))}();var n=r.O(void 0,[998],(function(){return r(9305)}));n=r.O(n)})();
//# sourceMappingURL=app.efb3f865.js.map