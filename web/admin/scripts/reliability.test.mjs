import assert from 'node:assert/strict'
import { test } from 'node:test'
import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
const server=await createServer({root:fileURLToPath(new URL('../',import.meta.url)),configFile:false,optimizeDeps:{noDiscovery:true,include:[]},server:{middlewareMode:true,hmr:false},appType:'custom'})
try{
 const {randomUUID,validPassword}=await server.ssrLoadModule('/src/lib/random.ts')
 const {closeTabs}=await server.ssrLoadModule('/src/layouts/tabs.ts')
 await test('UUID fallback and password validation work without secure-context randomUUID',()=>{const original=crypto.randomUUID;Object.defineProperty(crypto,'randomUUID',{value:undefined,configurable:true});try{assert.match(randomUUID(),/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);assert.equal(validPassword('中'.repeat(25)),false);assert.equal(validPassword('中'.repeat(4)),true)}finally{Object.defineProperty(crypto,'randomUUID',{value:original,configurable:true})}})
 await test('closing a background tab preserves an active page outside the tab list',()=>{assert.equal(closeTabs(['/','/a','/b'],'/forbidden','/a','close').active,'/forbidden')})
 for(const [module,name] of [['/src/composables/use-list.ts','useList'],['/src/features/system/files/use-resource-list.ts','useResourceList']]){
  const exports=await server.ssrLoadModule(module)
  await test(`${name} coalesces pagination/search and retains page/count after failure`,async()=>{
   let calls=0,resolve;const pending=new Promise(done=>{resolve=done});let state
   const app=createSSRApp({setup(){state=exports[name](async()=>{calls++;return pending});return()=>null}});await renderToString(app)
   state.query.page=3;const first=state.load();const duplicate=state.load();assert.equal(calls,1);resolve({items:[1],total:100});await Promise.all([first,duplicate]);assert.equal(state.query.page,3)
   let broken
   await renderToString(createSSRApp({setup(){broken=exports[name](async()=>{throw new Error('unavailable')});return()=>null}}));broken.query.page=3;broken.total.value=100;await broken.load();assert.equal(broken.query.page,3);assert.equal(broken.total.value,100);assert.ok(broken.error.value)
  })
 }
}finally{await server.close()}
