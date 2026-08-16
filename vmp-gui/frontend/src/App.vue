<template>
  <div class="h-screen w-screen flex flex-col bg-slate-100 border border-slate-300 text-slate-800 font-sans select-none overflow-hidden rounded shadow-sm">
    
    <!-- Custom Frameless Titlebar -->
    <div style="--wails-draggable:drag" class="h-10 flex items-center justify-between px-4 bg-slate-50 border-b border-slate-200 relative z-50 shrink-0">
      <div class="flex items-center space-x-2">
        <el-icon :size="16" color="#3b82f6"><Monitor /></el-icon>
        <span class="text-xs font-semibold text-slate-700 tracking-wide">
          VMProtect <span class="text-slate-500 font-normal ml-1">工作空间</span>
          <span v-if="currentFile" class="text-blue-600 ml-2">- {{ currentFile }}</span>
        </span>
      </div>
      <div style="--wails-draggable:no-drag" class="flex items-center space-x-0 h-full -mr-4">
        <button @click="minWindow" class="h-full px-4 hover:bg-slate-200 text-slate-500 hover:text-slate-800 transition-colors flex items-center justify-center">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="1" y="5" width="10" height="1" fill="currentColor"/></svg>
        </button>
        <button @click="maxWindow" class="h-full px-4 hover:bg-slate-200 text-slate-500 hover:text-slate-800 transition-colors flex items-center justify-center">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="1.5" y="1.5" width="9" height="9" stroke="currentColor" stroke-width="1"/></svg>
        </button>
        <button @click="closeWindow" class="h-full px-4 hover:bg-red-500 text-slate-500 hover:text-white transition-colors flex items-center justify-center rounded-tr">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M1 1L11 11M11 1L1 11" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
        </button>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="h-12 flex items-center px-4 bg-slate-100 border-b border-slate-200 space-x-2 shrink-0">
      <el-button color="#f8fafc" size="small" :icon="Folder" @click="openFile" class="text-slate-700 border-slate-300 hover:text-blue-600 hover:border-blue-400">打开目标文件</el-button>
      <div class="w-px h-5 bg-slate-300 mx-2"></div>
      <el-button color="#3b82f6" style="color: white;" size="small" :icon="VideoPlay" @click="runPacker" :disabled="!currentFile || selectedCount === 0" class="font-medium border-blue-500 hover:bg-blue-600 focus:outline-none transition-colors">开始保护</el-button>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex overflow-hidden bg-slate-100/50">
      
      <!-- Detail Sidebar (Project Tree) -->
      <div v-if="currentFile" class="w-64 bg-slate-50 border-r border-slate-200 flex flex-col h-full shrink-0">
        <div class="px-5 py-3 flex justify-between items-center bg-slate-100/50 border-b border-slate-200/50">
          <span class="text-[11px] font-bold text-slate-500 uppercase tracking-widest">项目元素</span>
          <el-tooltip content="关闭当前项目并返回主页" placement="right">
            <button @click="closeProject" class="text-slate-400 hover:text-blue-500 transition-colors p-1 rounded hover:bg-slate-200">
               <el-icon><House /></el-icon>
            </button>
          </el-tooltip>
        </div>
        <ul class="text-sm pt-2 flex-1 text-slate-600 font-medium">
          <li @click="currentTab = 'functions'" :class="{'bg-blue-50/50 text-blue-600 border-blue-500 shadow-[inset_1px_0_0_transparent]': currentTab === 'functions', 'hover:bg-slate-100 border-transparent': currentTab !== 'functions'}" class="px-5 py-2.5 cursor-pointer flex items-center border-r-2 transition-colors">
            <el-icon class="mr-3 text-lg" :class="{'text-blue-600': currentTab === 'functions', 'text-slate-400': currentTab !== 'functions'}"><StarFilled /></el-icon>目标函数
          </li>
          <li @click="currentTab = 'packed'" :class="{'bg-blue-50/50 text-blue-600 border-blue-500 shadow-[inset_1px_0_0_transparent]': currentTab === 'packed', 'hover:bg-slate-100 border-transparent': currentTab !== 'packed'}" class="px-5 py-2.5 cursor-pointer flex items-center border-r-2 transition-colors">
            <el-icon class="mr-3 text-lg" :class="{'text-blue-600': currentTab === 'packed', 'text-slate-400': currentTab !== 'packed'}"><Files /></el-icon>导出输出
          </li>
          <li @click="currentTab = 'options'" :class="{'bg-blue-50/50 text-blue-600 border-blue-500 shadow-[inset_1px_0_0_transparent]': currentTab === 'options', 'hover:bg-slate-100 border-transparent': currentTab !== 'options'}" class="px-5 py-2.5 cursor-pointer flex items-center border-r-2 transition-colors">
            <el-icon class="mr-3 text-lg" :class="{'text-blue-600': currentTab === 'options', 'text-slate-400': currentTab !== 'options'}"><Setting /></el-icon>保护选项
          </li>
        </ul>
      </div>

      <!-- Right Main Panel -->
      <div class="flex-1 flex flex-col h-full relative">
        
        <!-- Welcome View -->
        <div v-if="!currentFile" class="flex-1 flex h-full">
          
          <!-- Home Sidebar (Icon Only) -->
          <div class="w-14 bg-slate-50 border-r border-slate-200 flex flex-col items-center h-full shrink-0">
            <div class="w-full flex justify-center items-center bg-slate-100/50 border-b border-slate-200/50 h-12">
              <el-icon class="text-slate-400 text-lg"><Monitor /></el-icon>
            </div>
            <ul class="w-full text-sm pt-4 flex flex-col items-center flex-1 text-slate-600 font-medium space-y-2">
              <el-tooltip content="起始页" placement="right" :hide-after="0" effect="dark">
                <li @click="currentHomeTab = 'home'" :class="{'bg-blue-50/50 text-blue-600 border-blue-500 shadow-[inset_3px_0_0_transparent]': currentHomeTab === 'home', 'hover:bg-slate-100 border-transparent': currentHomeTab !== 'home'}" class="w-full h-12 flex justify-center items-center cursor-pointer border-l-2 transition-colors">
                  <el-icon class="text-xl h-full w-full" :class="{'text-blue-600': currentHomeTab === 'home', 'text-slate-400': currentHomeTab !== 'home'}"><House /></el-icon>
                </li>
              </el-tooltip>
              <el-tooltip content="更新日志" placement="right" :hide-after="0" effect="dark">
                <li @click="currentHomeTab = 'logs'" :class="{'bg-blue-50/50 text-blue-600 border-blue-500 shadow-[inset_3px_0_0_transparent]': currentHomeTab === 'logs', 'hover:bg-slate-100 border-transparent': currentHomeTab !== 'logs'}" class="w-full h-12 flex justify-center items-center cursor-pointer border-l-2 transition-colors">
                  <el-icon class="text-xl h-full w-full" :class="{'text-blue-600': currentHomeTab === 'logs', 'text-slate-400': currentHomeTab !== 'logs'}"><Document /></el-icon>
                </li>
              </el-tooltip>
            </ul>
          </div>

          <div class="flex-1 flex flex-col h-full bg-slate-100/50 relative">
            
            <!-- Home Contents -->
            <div v-if="currentHomeTab === 'home'" class="flex-1 flex items-start justify-center p-12 overflow-y-auto">
              <div class="w-full max-w-4xl bg-white border border-slate-200 text-slate-800 flex flex-col rounded-xl shadow-sm">
                <!-- Header -->
                <div class="h-32 bg-slate-50 border-b border-slate-100 flex items-center px-10 shrink-0 rounded-t-xl">
                   <div class="w-16 h-16 bg-blue-500/10 rounded-2xl flex items-center justify-center text-blue-500 shadow-[inset_0_2px_4px_rgba(0,0,0,0.05)] mr-6 border border-blue-100/50">
                     <el-icon :size="32"><Monitor /></el-icon>
                   </div>
                   <div class="flex flex-col">
                     <span class="text-2xl font-light tracking-tight text-slate-800">虚拟机保护系统 <span class="font-semibold text-blue-600/90">Core Engine</span></span>
                     <span class="text-sm text-slate-500 mt-1">准备就绪，请打开目标程序以开始操作。</span>
                   </div>
                </div>
                <!-- Body: Drop zone + Quick Actions -->
                <div class="flex flex-col p-8 items-stretch">
                   <!-- Drop Zone -->
                   <div
                     style="--wails-drop-target: drop"
                     @dragover.prevent="isDragOver = true"
                     @dragleave.prevent="isDragOver = false"
                     @drop.prevent="isDragOver = false"
                     :class="[
                       'border-2 border-dashed rounded-xl p-8 mb-6 flex flex-col items-center justify-center transition-all cursor-pointer',
                       isDragOver
                         ? 'border-blue-400 bg-blue-50/80 scale-[1.01]'
                         : 'border-slate-200 bg-slate-50/50 hover:border-blue-300 hover:bg-blue-50/30'
                     ]"
                     @click="openFile"
                   >
                     <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-4"
                       :class="isDragOver ? 'bg-blue-100 text-blue-500' : 'bg-slate-100 text-slate-400'">
                       <el-icon :size="28"><FolderOpened /></el-icon>
                     </div>
                     <span class="text-sm font-medium" :class="isDragOver ? 'text-blue-600' : 'text-slate-600'">
                       {{ isDragOver ? '松开以打开文件' : '拖拽 ELF 文件到此处，或点击选择' }}
                     </span>
                     <span class="text-xs text-slate-400 mt-2">支持 ARM64 ELF 可执行文件</span>
                   </div>

                   <div class="flex items-stretch">
                   <!-- Left Side: Quick actions -->
                   <div class="flex-1 pr-8">
                     <div class="text-[11px] text-slate-400 mb-4 font-bold uppercase tracking-wider">快速操作</div>
                     <ul class="text-sm space-y-2">
                       <li class="flex items-center cursor-pointer px-4 py-3 rounded-lg text-slate-700 hover:text-blue-600 bg-slate-50 hover:bg-blue-50 transition-colors font-medium border border-slate-100 hover:border-blue-200" @click="openFile">
                         <div class="w-8 h-8 rounded bg-white flex items-center justify-center mr-4 shadow-sm text-blue-500 border border-slate-100"><el-icon :size="16"><FolderOpened /></el-icon></div>
                         打开目标文件...
                       </li>
                       <li class="flex items-center cursor-pointer px-4 py-3 rounded-lg text-slate-700 hover:text-blue-600 bg-slate-50 hover:bg-blue-50 transition-colors border border-slate-100 hover:border-blue-200 font-medium">
                         <div class="w-8 h-8 rounded bg-white flex items-center justify-center mr-4 shadow-sm text-slate-400 border border-slate-100"><el-icon :size="16"><QuestionFilled /></el-icon></div>
                         查看文档 / 常见问题
                       </li>
                     </ul>
                   </div>
                   <!-- Right Side: Recent files -->
                   <div class="w-64 border-l border-slate-100 pl-8">
                     <div class="text-[11px] text-slate-400 mb-4 font-bold uppercase tracking-wider">最近打开文件</div>
                     <ul v-if="recentFiles.length > 0" class="text-sm space-y-1 text-slate-600">
                       <li v-for="rf in recentFiles" :key="rf.path" @click="loadFile(rf.path)" class="hover:bg-slate-50 hover:text-blue-600 rounded-md px-3 py-2 cursor-pointer flex items-center transition-colors" :title="rf.path">
                         <el-icon class="mr-3 text-slate-400 shrink-0"><Document /></el-icon>
                         <span class="truncate">{{ rf.name }}</span>
                       </li>
                     </ul>
                     <p v-else class="text-xs text-slate-400 italic px-3">暂无记录</p>
                   </div>
                   </div>
                </div>
              </div>
            </div>

            <!-- Logs Contents -->
            <div v-if="currentHomeTab === 'logs'" class="flex-1 flex items-start justify-center p-12 overflow-y-auto">
              <div class="w-full max-w-4xl bg-white border border-slate-200 text-slate-800 flex flex-col rounded-xl shadow-sm pb-8">
                 <div class="h-20 bg-slate-50 border-b border-slate-100 flex items-center px-10 shrink-0 rounded-t-xl mb-6">
                   <span class="text-lg font-medium text-slate-800 flex items-center"><el-icon class="mr-3 text-blue-500"><Document /></el-icon>系统更新记录 / Release Notes</span>
                 </div>
                 <div class="px-10 space-y-8">
                   <!-- Log Entry -->
                   <div class="flex items-start">
                     <div class="w-24 text-xs font-mono text-slate-400 pt-1 border-r border-slate-100 mr-6">v1.1.0</div>
                     <div class="flex-1">
                       <h4 class="text-sm font-semibold text-slate-800 mb-2">全新交互框架适配</h4>
                       <ul class="text-xs text-slate-600 space-y-1.5 list-disc list-inside">
                         <li>优化界面布局，新增全局侧边栏导航，主页与日志分离。</li>
                         <li>支持解析多函数列表以及函数复选、部分选定。</li>
                         <li>全界面中文化，符合目标操作习惯，修正全屏模式下的界面留白问题。</li>
                       </ul>
                     </div>
                   </div>
                   
                   <!-- Log Entry -->
                   <div class="flex items-start">
                     <div class="w-24 text-xs font-mono text-slate-400 pt-1 border-r border-slate-100 mr-6">v1.0.5</div>
                     <div class="flex-1">
                       <h4 class="text-sm font-semibold text-slate-800 mb-2">ARM64 底层引擎优化</h4>
                       <ul class="text-xs text-slate-600 space-y-1.5 list-disc list-inside">
                         <li>修复在 `-scan` 模式下存在的段错误跳出隐患。</li>
                         <li>支持多种复杂算法逻辑的虚拟化保护机制。</li>
                       </ul>
                     </div>
                   </div>

                   <!-- Log Entry -->
                   <div class="flex items-start opacity-70">
                     <div class="w-24 text-xs font-mono text-slate-400 pt-1 border-r border-slate-100 mr-6">v1.0.0</div>
                     <div class="flex-1">
                       <h4 class="text-sm font-semibold text-slate-800 mb-2">核心版本发布</h4>
                       <ul class="text-xs text-slate-600 space-y-1.5 list-disc list-inside">
                         <li>支持基础代码虚拟化和控制流混淆。</li>
                         <li>接入基于 Wails 的前端仪表盘，提供实时保护日志流。</li>
                       </ul>
                     </div>
                   </div>
                 </div>
              </div>
            </div>

          </div>
        </div>

        <!-- Working View -->
        <div v-else class="flex flex-col flex-1 h-full min-h-0 bg-slate-100/50" v-loading="isParsing" element-loading-text="正在解析节区与符号表..." element-loading-background="rgba(241, 245, 249, 0.8)">
           
           <!-- Tab 1: Functions Table -->
           <div v-show="currentTab === 'functions'" class="flex-1 flex flex-col min-h-0">
             <div class="h-12 border-b border-slate-200 flex items-center px-6 justify-between bg-white/60 backdrop-blur-sm shrink-0">
               <div class="flex items-center space-x-4">
                 <el-button color="#f8fafc" size="small" :icon="Plus" @click="showAddDialog = true" class="text-slate-700 border-slate-300 hover:text-blue-600 hover:border-blue-400">添加函数</el-button>
                 <span class="text-xs text-slate-500 whitespace-nowrap">
                   可保护 <strong class="font-semibold text-emerald-600">{{ compatibleCount }}</strong>
                   <span class="mx-1.5 text-slate-300">·</span>
                   已跳过 <strong class="font-semibold text-amber-600">{{ incompatibleCount }}</strong>
                 </span>
               </div>
               <div class="flex items-center space-x-4">
                 <div v-if="filteredAndSortedFunctions.length > pageSize" class="flex items-center space-x-3">
                   <span class="text-xs text-slate-500 whitespace-nowrap">共 {{ filteredAndSortedFunctions.length }} 项</span>
                   <el-pagination
                     v-model:current-page="currentPage"
                     :page-size="pageSize"
                     :total="filteredAndSortedFunctions.length"
                     layout="prev, pager, next"
                     small
                     background
                   />
                 </div>
                 <el-input v-model="searchQuery" size="small" placeholder="搜索函数..." class="light-input w-72">
                   <template #prefix><el-icon><Search /></el-icon></template>
                 </el-input>
               </div>
             </div>

             <div class="flex-1 overflow-auto p-6 flex justify-center items-start">
               <div class="w-full max-w-7xl bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden pb-4">
                 <table class="w-full text-sm text-left">
                   <thead class="bg-slate-50 border-b border-slate-200 text-slate-500 text-xs uppercase font-semibold tracking-wider">
                     <tr>
                       <th class="py-3 px-6 w-12 text-center">
                         <input
                           type="checkbox"
                           class="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                           title="选择全部可保护函数"
                           @change="toggleAll"
                           :checked="isAllSelected"
                           :indeterminate="isPartiallySelected"
                           :disabled="compatibleCount === 0"
                         />
                       </th>
                       <th class="py-3 px-6">函数节点</th>
                       <th class="py-3 px-6">地址</th>
                       <th class="py-3 px-6">大小</th>
                       <th class="py-3 px-6">来源</th>
                       <th class="py-3 px-6">保护状态</th>
                     </tr>
                   </thead>
                   <tbody class="divide-y divide-slate-100">
                     <tr
                       v-for="fn in pagedFunctions"
                       :key="`${fn.address}:${fn.name}`"
                       :class="[
                         'transition-colors group',
                         isCompatible(fn) ? 'hover:bg-blue-50/50 cursor-pointer' : 'bg-slate-50/70 text-slate-400 cursor-not-allowed'
                       ]"
                       @click="toggleFunction(fn)"
                     >
                       <td class="py-3 px-6 text-center" @click.stop>
                          <input
                            type="checkbox"
                            v-model="fn.selected"
                            :disabled="!isCompatible(fn)"
                            :title="isCompatible(fn) ? '选择此函数' : fn.compatibilityReason"
                            class="rounded border-slate-300 text-blue-600 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-40"
                          />
                       </td>
                       <td class="py-3 px-6 whitespace-nowrap text-slate-800 font-medium flex items-center">
                         <el-icon class="mr-3 text-slate-400 group-hover:text-blue-500 transition-colors"><Setting /></el-icon>{{ fn.name }}
                       </td>
                       <td class="py-3 px-6 text-slate-500 font-mono text-xs">{{ fn.address }}</td>
                       <td class="py-3 px-6 text-slate-500 font-mono text-xs">{{ fn.size }} B</td>
                       <td class="py-3 px-6">
                         <span class="inline-flex items-center px-2 py-1 rounded text-[11px] font-medium bg-slate-100 text-slate-600 border border-slate-200 whitespace-nowrap">
                           {{ fn.source || 'manual' }} · {{ fn.confidence || 'user' }}
                         </span>
                       </td>
                       <td class="py-3 px-6">
                          <el-tooltip v-if="!isCompatible(fn)" :content="fn.compatibilityReason || '包含当前不支持的 SIMD/FPU 指令'" placement="left">
                            <span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200 whitespace-nowrap">
                              已跳过（SIMD）
                            </span>
                          </el-tooltip>
                          <span v-else-if="fn.selected" class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-semibold bg-blue-100 text-blue-700 border border-blue-200">
                            待保护 ({{ fn.protection }})
                          </span>
                          <span v-else class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-semibold bg-slate-100 text-slate-500 border border-slate-200">
                            未选择
                          </span>
                       </td>
                     </tr>
                     <tr v-if="filteredAndSortedFunctions.length === 0">
                       <td colspan="6" class="py-12 text-center text-slate-400 italic">未自动发现函数范围，可使用“添加函数”输入已确认的地址范围。</td>
                     </tr>
                   </tbody>
                 </table>
               </div>
             </div>
           </div>

           <!-- Tab 2: Packed Output settings -->
           <div v-show="currentTab === 'packed'" class="flex-1 overflow-auto p-12 flex justify-center items-start">
              <div class="w-full max-w-7xl bg-white border border-slate-200 rounded-xl p-10 shadow-sm">
                <h2 class="text-lg font-semibold text-slate-800 mb-8 flex items-center border-b border-slate-100 pb-4">
                  <el-icon class="mr-3 text-blue-500 text-xl"><Files /></el-icon> 导出输出配置
                </h2>
                
                <div class="space-y-6">
                  <div>
                    <label class="block text-sm font-medium text-slate-700 mb-3">指定保存路径 (Output Path)</label>
                    <div class="flex space-x-3">
                      <el-input v-model="outputPath" placeholder="默认保存到原文件目录下的 _vmp 副本" class="flex-1 custom-input" size="large"></el-input>
                      <el-button color="#f8fafc" class="ml-2 text-slate-700 border-slate-300 h-10 px-6" @click="browseOutputPath">浏览...</el-button>
                    </div>
                    <p class="mt-3 text-xs text-slate-500 flex items-center">
                      <el-icon class="mr-1"><InfoFilled /></el-icon>保护打包完成后的可执行文件将保存到此位置。
                    </p>
                  </div>

                  <div v-if="enableDebug">
                    <label class="block text-sm font-medium text-slate-700 mb-3">Debug 对照文件路径</label>
                    <div class="flex space-x-3">
                      <el-input v-model="debugFilePath" class="flex-1 custom-input" size="large"></el-input>
                      <el-button color="#f8fafc" class="ml-2 text-slate-700 border-slate-300 h-10 px-6" @click="browsePath('debug')">浏览...</el-button>
                    </div>
                    <p class="mt-3 text-xs text-slate-500 flex items-center">
                      <el-icon class="mr-1"><InfoFilled /></el-icon>开启 Debug 模式后，虚拟机指令对照信息将写入此文件。
                    </p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-slate-700 mb-3">不支持指令 Debug 文件路径</label>
                    <div class="flex space-x-3">
                      <el-input v-model="unsupportedFilePath" class="flex-1 custom-input" size="large"></el-input>
                      <el-button color="#f8fafc" class="ml-2 text-slate-700 border-slate-300 h-10 px-6" @click="browsePath('unsupported')">浏览...</el-button>
                    </div>
                    <p class="mt-3 text-xs text-slate-500 flex items-center">
                      <el-icon class="mr-1"><InfoFilled /></el-icon>当遇到无法虚拟化的指令时，Packer 会将详细信息写入此文件。
                    </p>
                  </div>
                </div>
              </div>
           </div>

           <!-- Tab 3: Options -->
           <div v-show="currentTab === 'options'" class="flex-1 overflow-auto p-12 flex justify-center items-start">
              <div class="w-full max-w-7xl bg-white border border-slate-200 rounded-xl p-10 shadow-sm">
                <h2 class="text-lg font-semibold text-slate-800 mb-8 flex items-center border-b border-slate-100 pb-4">
                  <el-icon class="mr-3 text-blue-500 text-xl"><Setting /></el-icon> 保护选项设置
                </h2>
                
                <div class="space-y-8 divide-y divide-slate-100">
                  <div class="flex items-center justify-between pb-2">
                    <div class="pr-8">
                      <div class="text-sm font-medium text-slate-800 mb-1">开启 Debug 模式</div>
                      <div class="text-[13px] text-slate-500 leading-relaxed">打包后在运行时输出虚拟机调试指令流水线日志。通常用于分析执行流或排除故障。</div>
                    </div>
                    <div>
                      <el-switch v-model="enableDebug"></el-switch>
                    </div>
                  </div>
                  
                  <div class="flex items-center justify-between pt-6">
                    <div class="pr-8">
                      <div class="text-sm font-medium text-slate-800 mb-1">去除符号表 (Strip Symbols)</div>
                      <div class="text-[13px] text-slate-500 leading-relaxed">保护完成后移除文件中的所有调试符号和未导出函数名，增大逆向分析难度。推荐在发行版中开启。</div>
                    </div>
                    <div>
                      <el-switch v-model="stripSymbols"></el-switch>
                    </div>
                  </div>
                  
                  <div class="flex items-center justify-between pt-6">
                    <div class="pr-8">
                      <div class="text-sm font-medium text-slate-800 mb-1">Token 化入口模式</div>
                      <div class="text-[13px] text-slate-500 leading-relaxed">使用 3 指令精简跳板替代默认入口，减小代码体积占用。</div>
                    </div>
                    <div>
                      <el-switch v-model="tokenEntry"></el-switch>
                    </div>
                  </div>
                </div>
              </div>
           </div>

           <!-- Clean Log Console (Always visible at bottom when file loaded) -->
           <div class="h-48 border-t border-slate-200 bg-slate-50 flex flex-col relative shadow-[inset_0_2px_10px_rgba(0,0,0,0.02)] shrink-0">
             <div class="h-8 bg-slate-100 border-b border-slate-200 px-6 flex items-center justify-between">
               <span class="text-[10px] text-slate-500 font-bold uppercase tracking-widest flex items-center">
                 <span class="w-1.5 h-1.5 rounded-full bg-slate-400 mr-2"></span>
                 输出控制台
               </span>
             </div>
             <div class="flex-1 p-4 overflow-y-auto font-mono text-xs leading-relaxed text-slate-700 bg-white" id="terminal-output">
               <p v-if="logs.length === 0" class="text-slate-400">系统空闲。准备就绪...</p>
               <div v-for="(log, i) in logs" :key="i" class="flex mb-1">
                  <span class="text-slate-400 mr-3 select-none">[{{ new Date().toLocaleTimeString().split(' ')[0] }}]</span>
                  <span :class="{'text-red-500 font-medium': log.includes('[x]') || log.includes('[!]'), 'text-slate-800': log.includes('[+]') || log.includes('>') }">
                     {{ log }}
                  </span>
               </div>
             </div>
           </div>
        </div>

      </div>
    </div>
    
    <!-- Footer StatusBar -->
    <div class="h-7 bg-slate-800 flex items-center px-4 text-[11px] justify-between font-medium text-slate-300 tracking-wide border-t border-slate-700">
      <span class="flex items-center">
        <el-icon class="mr-2 text-blue-400" v-if="isProtecting"><VideoPlay class="animate-spin" /></el-icon>
        <el-icon class="mr-2 text-slate-400" v-else><Monitor /></el-icon>
        {{ currentFile ? `${currentPath}` : '准备就绪' }}
      </span>
      <span class="text-slate-500">核心引擎: v1.0.0</span>
    </div>

    <!-- Add Function Dialog -->
    <el-dialog v-model="showAddDialog" title="添加自定义函数" width="460" :close-on-click-modal="false" align-center destroy-on-close>
      <div style="display: flex; flex-direction: column; gap: 20px;">
        <div>
          <label style="display: block; font-size: 13px; font-weight: 500; color: #334155; margin-bottom: 8px;">方法名</label>
          <el-input v-model="newFuncForm.name" placeholder="例如: my_encrypt_func" size="large" />
        </div>
        <div>
          <label style="display: block; font-size: 13px; font-weight: 500; color: #334155; margin-bottom: 8px;">开始地址</label>
          <el-input v-model="newFuncForm.startAddress" placeholder="十六进制，例如: 401000" size="large" />
        </div>
        <div>
          <label style="display: block; font-size: 13px; font-weight: 500; color: #334155; margin-bottom: 8px;">结束地址</label>
          <el-input v-model="newFuncForm.endAddress" placeholder="十六进制，例如: 401100" size="large" />
        </div>
      </div>
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 8px;">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="addFunction" :disabled="!newFuncForm.name.trim() || !newFuncForm.startAddress.trim() || !newFuncForm.endAddress.trim()">确认添加</el-button>
        </div>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, watch, watchEffect, nextTick, computed } from 'vue'
import {
  Folder, Document, VideoPlay, Search,
  HelpFilled, Lock, InfoFilled, StarFilled, Key, Files, Connection, Setting,
  Monitor, FolderOpened, Reading, QuestionFilled, Plus, Promotion, Opportunity, House
} from '@element-plus/icons-vue'
import { SelectFile, AnalyzeELF, Protect, GetRecentFiles, AddRecentFile, SelectSaveFile } from '../wailsjs/go/api/VMPEngine'
import { WindowMinimise, WindowToggleMaximise, Quit, EventsOn } from '../wailsjs/runtime/runtime'

// State definitions
const currentFile = ref('')
const currentPath = ref('')
const currentFunctions = ref<any[]>([])
const logs = ref<string[]>([])
const isProtecting = ref(false)
const isParsing = ref(false)

// Tabs State
const currentHomeTab = ref('home') // 'home', 'logs'
const currentTab = ref('functions') // 'functions', 'packed', 'options'

// Settings State
const outputPath = ref('')
const debugFilePath = ref('')
const unsupportedFilePath = ref('')
const enableDebug = ref(false)
const stripSymbols = ref(false)
const tokenEntry = ref(false)
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 100
const showAddDialog = ref(false)
const newFuncForm = ref({ name: '', startAddress: '', endAddress: '' })
const recentFiles = ref<Array<{name: string, path: string}>>([])

// Load recent files on startup
const loadRecentFiles = async () => {
  try {
    const files = await GetRecentFiles()
    recentFiles.value = (files || []).map((f: any) => ({ name: f.name || '', path: f.path || '' }))
  } catch { recentFiles.value = [] }
}
loadRecentFiles()

const closeProject = () => {
  currentFile.value = ''
  currentPath.value = ''
  currentFunctions.value = []
  currentHomeTab.value = 'home'
}

// Select all helper capabilities
const isCompatible = (fn: any) => fn.compatible !== false

const selectableFunctions = computed(() => currentFunctions.value.filter(isCompatible))
const compatibleCount = computed(() => selectableFunctions.value.length)
const incompatibleCount = computed(() => currentFunctions.value.length - compatibleCount.value)
const selectedCount = computed(() => selectableFunctions.value.filter(fn => fn.selected).length)

const isAllSelected = computed(() => {
  return compatibleCount.value > 0 && selectedCount.value === compatibleCount.value
})

const isPartiallySelected = computed(() => selectedCount.value > 0 && !isAllSelected.value)

const filteredAndSortedFunctions = computed(() => {
  let list = currentFunctions.value
  if (searchQuery.value.trim()) {
    const keyword = searchQuery.value.toLowerCase()
    list = list.filter(fn =>
      fn.name.toLowerCase().includes(keyword) ||
      fn.address.toLowerCase().includes(keyword) ||
      (fn.source || '').toLowerCase().includes(keyword) ||
      (fn.compatibilityReason || '').toLowerCase().includes(keyword)
    )
  }
  return [...list]
})

const pagedFunctions = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredAndSortedFunctions.value.slice(start, start + pageSize)
})

watch(searchQuery, () => {
  currentPage.value = 1
})

const toggleAll = (e: Event) => {
  const checked = (e.target as HTMLInputElement).checked
  currentFunctions.value.forEach(fn => {
    fn.selected = isCompatible(fn) ? checked : false
  })
}

const toggleFunction = (fn: any) => {
  if (isCompatible(fn)) fn.selected = !fn.selected
}

// Auto-scroll terminal
const scrollToBottom = () => {
  nextTick(() => {
    const el = document.getElementById('terminal-output');
    if (el) el.scrollTop = el.scrollHeight;
  })
};

watchEffect(() => {
  if (logs.value.length) scrollToBottom()
})

// Custom Window Controls
const minWindow = () => WindowMinimise()
const maxWindow = () => WindowToggleMaximise()
const closeWindow = () => Quit()

// Subscribe to Wails custom logs
EventsOn('vmp-log', (msg) => {
  logs.value.push(msg)
})

// Subscribe to file drop events
const isDragOver = ref(false)
EventsOn('wails:file-drop', async (_x: number, _y: number, paths: string[]) => {
  isDragOver.value = false
  if (!paths || paths.length === 0) return
  const filePath = paths[0]
  await loadFile(filePath)
})

import { ElMessage } from 'element-plus'

const loadFile = async (selectedPath: string) => {
  try {
     const segments = selectedPath.split(/[/\\]/);
     currentFile.value = segments[segments.length - 1] || 'Loading...'
     currentPath.value = selectedPath
     isParsing.value = true
     currentTab.value = 'functions'
     currentFunctions.value = []
     logs.value.push(`[+] 正在加载目标文件: ${selectedPath}`)
     
     const result: any = await AnalyzeELF(selectedPath)
     
     currentFile.value = result.fileName
     currentPath.value = result.filePath
     outputPath.value = result.filePath + "_vmp"
     debugFilePath.value = result.filePath + "_vmp.debug.txt"
     unsupportedFilePath.value = result.filePath + "_vmp.unsupported.txt"
     
     currentFunctions.value = (result.functions || []).map((f: any) => ({
       ...f,
       compatible: f.compatible !== false,
       selected: false
     }))
     currentPage.value = 1
     
     logs.value.push(`[+] 架构: ${result.arch}, 格式: ${result.format}`)
     logs.value.push(`[+] 函数范围分析完毕，共 ${currentFunctions.value.length} 个函数；可保护 ${compatibleCount.value} 个，已跳过 SIMD/FPU 函数 ${incompatibleCount.value} 个。`)
     for (const warning of (result.warnings || [])) {
       logs.value.push(`[!] ${warning}`)
     }
     
     // Save to recent files
     AddRecentFile(selectedPath).then(() => loadRecentFiles())
  } catch (err: any) {
     logs.value.push(`[!] 读取文件失败: ${err}`)
     ElMessage.error(`打开失败: ${err}`)
     currentFile.value = ''
     currentPath.value = ''
  } finally {
     isParsing.value = false
  }
}

const browseOutputPath = async () => {
  try {
    const defaultName = currentFile.value ? currentFile.value + '_vmp' : ''
    const selected = await SelectSaveFile(defaultName)
    if (selected) {
      outputPath.value = selected
    }
  } catch (err: any) {
    ElMessage.error(`选择路径失败: ${err}`)
  }
}

const browsePath = async (type: string) => {
  try {
    const defaultName = type === 'debug'
      ? (currentFile.value ? currentFile.value + '_vmp.debug.txt' : 'debug.txt')
      : (currentFile.value ? currentFile.value + '_vmp.unsupported.txt' : 'unsupported.txt')
    const selected = await SelectSaveFile(defaultName)
    if (selected) {
      if (type === 'debug') {
        debugFilePath.value = selected
      } else {
        unsupportedFilePath.value = selected
      }
    }
  } catch (err: any) {
    ElMessage.error(`选择路径失败: ${err}`)
  }
}

const openFile = async () => {
  try {
     const selectedPath = await SelectFile()
     if (!selectedPath) {
        return // User cancelled
     }
     await loadFile(selectedPath)
  } catch (err: any) {
     logs.value.push(`[!] 选择文件失败: ${err}`)
     ElMessage.error(`打开失败: ${err}`)
  }
}

const runPacker = async () => {
  const targets = currentFunctions.value.filter(fn => fn.selected && isCompatible(fn))
  if (!currentPath.value || isProtecting.value || targets.length === 0) return;
  
  isProtecting.value = true
  logs.value = [] // clear logs
  
  try {
    logs.value.push(`[+] 准备开始保护过程，选中 ${targets.length} 个函数节点...`)
    logs.value.push(`[+] 设置 - Debug模式: ${enableDebug.value ? '启用' : '禁用'}, 去除符号表: ${stripSymbols.value ? '启用' : '禁用'}`)
    logs.value.push(`[+] 输出路径: ${outputPath.value}`)
    await Protect({
      "file": currentPath.value,
      "functions": targets,
      "options": {
        "outputPath": outputPath.value,
        "debug": enableDebug.value,
        "stripSymbols": stripSymbols.value,
        "tokenEntry": tokenEntry.value
      }
    })
    logs.value.push(`[+] 操作全部完成。`)
  } catch(err: any){
    logs.value.push(`[x] 执行失败: ${err}`)
  } finally {
    isProtecting.value = false
  }
}

const addFunction = () => {
  const name = newFuncForm.value.name.trim()
  let startHex = newFuncForm.value.startAddress.trim().replace(/^0x/i, '')
  let endHex = newFuncForm.value.endAddress.trim().replace(/^0x/i, '')

  if (!name || !startHex || !endHex) {
    ElMessage.warning('请填写完整的函数信息')
    return
  }

  const startAddr = parseInt(startHex, 16)
  const endAddr = parseInt(endHex, 16)

  if (isNaN(startAddr) || isNaN(endAddr)) {
    ElMessage.error('地址格式无效，请输入合法的十六进制地址')
    return
  }

  if (endAddr <= startAddr) {
    ElMessage.error('结束地址必须大于开始地址')
    return
  }

  const size = endAddr - startAddr

  currentFunctions.value.push({
    name: name,
    address: '0x' + startAddr.toString(16).toUpperCase(),
    size: size,
    source: 'manual',
    confidence: 'user',
    protection: 'Virtualization',
    compatible: true,
    compatibilityReason: '',
    selected: true,
    isCustom: true
  })
  searchQuery.value = ''
  currentPage.value = Math.ceil(currentFunctions.value.length / pageSize)

  logs.value.push(`[+] 已手动添加函数: ${name} @ 0x${startAddr.toString(16).toUpperCase()} - 0x${endAddr.toString(16).toUpperCase()} (${size} bytes)`)
  ElMessage.success(`函数 ${name} 已添加`)

  // Reset form
  newFuncForm.value = { name: '', startAddress: '', endAddress: '' }
  showAddDialog.value = false
}
</script>

<style scoped>
/* Inject some specific theme overrides for element-plus if needed */
:deep(.el-input__wrapper) {
  background-color: rgb(248 250 252 / 0.5) !important;
  border-color: rgb(226 232 240) !important;
  box-shadow: none !important;
  border-radius: 6px;
}
:deep(.el-input__inner) {
  color: rgb(51 65 85) !important;
}
:deep(.el-input__inner::placeholder) {
  color: rgb(148 163 184) !important;
}
:deep(.el-icon) {
  color: inherit;
}
/* Override Element Plus switch colors for standard styling */
:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #3b82f6 !important;
  border-color: #3b82f6 !important;
}
</style>
