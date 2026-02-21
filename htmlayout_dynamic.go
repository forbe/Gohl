package gohl

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	htmlayoutLib *syscall.DLL
)

func init() {
	// 初始化资源目录
	extractResources()
	// 初始化 htmlayout 库
	initHtmlayoutLib()
}

func initHtmlayoutLib() {
	if resourcesDir == "" {
		fmt.Println("资源目录未初始化")
		return
	}

	dllPath := filepath.Join(resourcesDir, "htmlayout.dll")

	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		fmt.Printf("htmlayout.dll 不存在: %s\n", dllPath)
		return
	}

	var err error
	htmlayoutLib, err = syscall.LoadDLL(dllPath)
	if err != nil {
		fmt.Printf("加载 htmlayout.dll 失败: %v\n", err)
		return
	}

	fmt.Printf("成功加载 htmlayout.dll: %s\n", dllPath)

	initHtmlayoutFunctions()
}

var (
	procHTMLayoutProcND                   *syscall.Proc
	procHTMLayoutLoadHtmlEx               *syscall.Proc
	procHTMLayoutLoadFile                 *syscall.Proc
	procHTMLayoutSetOption                *syscall.Proc
	procHTMLayoutDataReady                *syscall.Proc
	procHTMLayoutWindowAttachEventHandler *syscall.Proc
	procHTMLayoutWindowDetachEventHandler *syscall.Proc
	procHTMLayoutSetCallback              *syscall.Proc
	procHTMLayoutGetRootElement           *syscall.Proc
	procHTMLayoutCreateElement            *syscall.Proc
	procHTMLayoutFindElement              *syscall.Proc
	procHTMLayoutGetFocusElement          *syscall.Proc
	procHTMLayoutAttachEventHandler       *syscall.Proc
	procHTMLayoutAttachEventHandlerEx     *syscall.Proc
	procHTMLayoutDetachEventHandler       *syscall.Proc
	procHTMLayoutUpdateElementEx          *syscall.Proc
	procHTMLayoutSetCapture               *syscall.Proc
	procHTMLayoutShowPopup                *syscall.Proc
	procHTMLayoutShowPopupAt              *syscall.Proc
	procHTMLayoutHidePopup                *syscall.Proc
	procHTMLayoutSelectElements           *syscall.Proc
	procHTMLayoutSelectParent             *syscall.Proc
	procHTMLayoutSendEvent                *syscall.Proc
	procHTMLayoutPostEvent                *syscall.Proc
	procHTMLayoutGetChildrenCount         *syscall.Proc
	procHTMLayoutGetNthChild              *syscall.Proc
	procHTMLayoutGetElementIndex          *syscall.Proc
	procHTMLayoutGetParentElement         *syscall.Proc
	procHTMLayoutInsertElement            *syscall.Proc
	procHTMLayoutDetachElement            *syscall.Proc
	procHTMLayoutDeleteElement            *syscall.Proc
	procHTMLayoutCloneElement             *syscall.Proc
	procHTMLayoutSwapElements             *syscall.Proc
	procHTMLayoutSetEventRoot             *syscall.Proc
	procHTMLayoutScrollToView             *syscall.Proc
	procHTMLayoutGetElementUID            *syscall.Proc
	procHTMLayoutGetElementByUID          *syscall.Proc
	procHTMLayoutCallBehaviorMethod       *syscall.Proc
	procHTMLayoutCombineURL               *syscall.Proc
	procHTMLayoutSortElements             *syscall.Proc
	procHTMLayoutSetTimerEx               *syscall.Proc
	procHTMLayoutGetElementHwnd           *syscall.Proc
	procHTMLayoutGetElementHtml           *syscall.Proc
	procHTMLayoutGetElementType           *syscall.Proc
	procHTMLayoutSetElementHtml           *syscall.Proc
	procHTMLayoutSetElementInnerText      *syscall.Proc
	procHTMLayoutGetElementInnerText      *syscall.Proc
	procHTMLayoutGetAttributeByName       *syscall.Proc
	procHTMLayoutSetAttributeByName       *syscall.Proc
	procHTMLayoutGetNthAttribute          *syscall.Proc
	procHTMLayoutGetAttributeCount        *syscall.Proc
	procHTMLayoutGetStyleAttribute        *syscall.Proc
	procHTMLayoutSetStyleAttribute        *syscall.Proc
	procHTMLayoutGetElementState          *syscall.Proc
	procHTMLayoutSetElementState          *syscall.Proc
	procHTMLayoutMoveElement              *syscall.Proc
	procHTMLayoutMoveElementEx            *syscall.Proc
	procHTMLayoutGetElementLocation       *syscall.Proc
	procHTMLayout_UseElement              *syscall.Proc
	procHTMLayout_UnuseElement            *syscall.Proc
)

func initHtmlayoutFunctions() {
	if htmlayoutLib == nil {
		return
	}

	procHTMLayoutProcND = mustFindProc("HTMLayoutProcND")
	procHTMLayoutLoadHtmlEx = mustFindProc("HTMLayoutLoadHtmlEx")
	procHTMLayoutLoadFile = mustFindProc("HTMLayoutLoadFile")
	procHTMLayoutSetOption = mustFindProc("HTMLayoutSetOption")
	procHTMLayoutDataReady = mustFindProc("HTMLayoutDataReady")
	procHTMLayoutWindowAttachEventHandler = mustFindProc("HTMLayoutWindowAttachEventHandler")
	procHTMLayoutWindowDetachEventHandler = mustFindProc("HTMLayoutWindowDetachEventHandler")
	procHTMLayoutSetCallback = mustFindProc("HTMLayoutSetCallback")
	procHTMLayoutGetRootElement = mustFindProc("HTMLayoutGetRootElement")
	procHTMLayoutCreateElement = mustFindProc("HTMLayoutCreateElement")
	procHTMLayoutFindElement = mustFindProc("HTMLayoutFindElement")
	procHTMLayoutGetFocusElement = mustFindProc("HTMLayoutGetFocusElement")
	procHTMLayoutAttachEventHandler = mustFindProc("HTMLayoutAttachEventHandler")
	procHTMLayoutAttachEventHandlerEx = mustFindProc("HTMLayoutAttachEventHandlerEx")
	procHTMLayoutDetachEventHandler = mustFindProc("HTMLayoutDetachEventHandler")
	procHTMLayoutUpdateElementEx = mustFindProc("HTMLayoutUpdateElementEx")
	procHTMLayoutSetCapture = mustFindProc("HTMLayoutSetCapture")
	procHTMLayoutShowPopup = mustFindProc("HTMLayoutShowPopup")
	procHTMLayoutShowPopupAt = mustFindProc("HTMLayoutShowPopupAt")
	procHTMLayoutHidePopup = mustFindProc("HTMLayoutHidePopup")
	procHTMLayoutSelectElements = mustFindProc("HTMLayoutSelectElements")
	procHTMLayoutSelectParent = mustFindProc("HTMLayoutSelectParent")
	procHTMLayoutSendEvent = mustFindProc("HTMLayoutSendEvent")
	procHTMLayoutPostEvent = mustFindProc("HTMLayoutPostEvent")
	procHTMLayoutGetChildrenCount = mustFindProc("HTMLayoutGetChildrenCount")
	procHTMLayoutGetNthChild = mustFindProc("HTMLayoutGetNthChild")
	procHTMLayoutGetElementIndex = mustFindProc("HTMLayoutGetElementIndex")
	procHTMLayoutGetParentElement = mustFindProc("HTMLayoutGetParentElement")
	procHTMLayoutInsertElement = mustFindProc("HTMLayoutInsertElement")
	procHTMLayoutDetachElement = mustFindProc("HTMLayoutDetachElement")
	procHTMLayoutDeleteElement = mustFindProc("HTMLayoutDeleteElement")
	procHTMLayoutCloneElement = mustFindProc("HTMLayoutCloneElement")
	procHTMLayoutSwapElements = mustFindProc("HTMLayoutSwapElements")
	procHTMLayoutSetEventRoot = mustFindProc("HTMLayoutSetEventRoot")
	procHTMLayoutScrollToView = mustFindProc("HTMLayoutScrollToView")
	procHTMLayoutGetElementUID = mustFindProc("HTMLayoutGetElementUID")
	procHTMLayoutGetElementByUID = mustFindProc("HTMLayoutGetElementByUID")
	procHTMLayoutCallBehaviorMethod = mustFindProc("HTMLayoutCallBehaviorMethod")
	procHTMLayoutCombineURL = mustFindProc("HTMLayoutCombineURL")
	procHTMLayoutSortElements = mustFindProc("HTMLayoutSortElements")
	procHTMLayoutSetTimerEx = mustFindProc("HTMLayoutSetTimerEx")
	procHTMLayoutGetElementHwnd = mustFindProc("HTMLayoutGetElementHwnd")
	procHTMLayoutGetElementHtml = mustFindProc("HTMLayoutGetElementHtml")
	procHTMLayoutGetElementType = mustFindProc("HTMLayoutGetElementType")
	procHTMLayoutSetElementHtml = mustFindProc("HTMLayoutSetElementHtml")
	procHTMLayoutSetElementInnerText = mustFindProc("HTMLayoutSetElementInnerText")
	procHTMLayoutGetElementInnerText = mustFindProc("HTMLayoutGetElementInnerText")
	procHTMLayoutGetAttributeByName = mustFindProc("HTMLayoutGetAttributeByName")
	procHTMLayoutSetAttributeByName = mustFindProc("HTMLayoutSetAttributeByName")
	procHTMLayoutGetNthAttribute = mustFindProc("HTMLayoutGetNthAttribute")
	procHTMLayoutGetAttributeCount = mustFindProc("HTMLayoutGetAttributeCount")
	procHTMLayoutGetStyleAttribute = mustFindProc("HTMLayoutGetStyleAttribute")
	procHTMLayoutSetStyleAttribute = mustFindProc("HTMLayoutSetStyleAttribute")
	procHTMLayoutGetElementState = mustFindProc("HTMLayoutGetElementState")
	procHTMLayoutSetElementState = mustFindProc("HTMLayoutSetElementState")
	procHTMLayoutMoveElement = mustFindProc("HTMLayoutMoveElement")
	procHTMLayoutMoveElementEx = mustFindProc("HTMLayoutMoveElementEx")
	procHTMLayoutGetElementLocation = mustFindProc("HTMLayoutGetElementLocation")
	procHTMLayout_UseElement = mustFindProc("HTMLayout_UseElement")
	procHTMLayout_UnuseElement = mustFindProc("HTMLayout_UnuseElement")
}

func mustFindProc(name string) *syscall.Proc {
	proc, err := htmlayoutLib.FindProc(name)
	if err != nil {
		fmt.Printf("找不到函数 %s: %v\n", name, err)
		return nil
	}
	// fmt.Printf("加载函数 %s: %v\n", name, proc.Addr())
	return proc
}

func HTMLayoutProcND(hwnd uintptr, msg uint32, wparam uintptr, lparam uintptr, handled *bool) int {
	if procHTMLayoutProcND == nil {
		return 0
	}
	ret, _, _ := procHTMLayoutProcND.Call(hwnd, uintptr(msg), wparam, lparam, uintptr(unsafe.Pointer(handled)))
	return int(ret)
}

func HTMLayoutLoadHtmlEx(hwnd uintptr, data *byte, dataSize uint32, baseUrl *uint16) bool {
	if procHTMLayoutLoadHtmlEx == nil {
		return false
	}
	ret, _, _ := procHTMLayoutLoadHtmlEx.Call(hwnd, uintptr(unsafe.Pointer(data)), uintptr(dataSize), uintptr(unsafe.Pointer(baseUrl)))
	return ret != 0
}

func HTMLayoutLoadFile(hwnd uintptr, uri *uint16) bool {
	if procHTMLayoutLoadFile == nil {
		return false
	}
	ret, _, _ := procHTMLayoutLoadFile.Call(hwnd, uintptr(unsafe.Pointer(uri)))
	return ret != 0
}

func HTMLayoutSetOption(hwnd uintptr, option uint32, value uint32) bool {
	if procHTMLayoutSetOption == nil {
		return false
	}
	ret, _, _ := procHTMLayoutSetOption.Call(hwnd, uintptr(option), uintptr(value))
	return ret != 0
}

func HTMLayoutDataReady(hwnd uintptr, uri *uint16, data *byte, dataSize uint32) bool {
	if procHTMLayoutDataReady == nil {
		return false
	}
	ret, _, _ := procHTMLayoutDataReady.Call(hwnd, uintptr(unsafe.Pointer(uri)), uintptr(unsafe.Pointer(data)), uintptr(dataSize))
	return ret != 0
}

func HTMLayoutWindowAttachEventHandler(hwnd uintptr, proc uintptr, tag uintptr, subscription uint32) int {
	if procHTMLayoutWindowAttachEventHandler == nil {
		return 0
	}
	ret, _, _ := procHTMLayoutWindowAttachEventHandler.Call(hwnd, proc, tag, uintptr(subscription))
	return int(ret)
}

func HTMLayoutWindowDetachEventHandler(hwnd uintptr, proc uintptr, tag uintptr) int {
	if procHTMLayoutWindowDetachEventHandler == nil {
		return 0
	}
	ret, _, _ := procHTMLayoutWindowDetachEventHandler.Call(hwnd, proc, tag)
	return int(ret)
}

func HTMLayoutSetCallback(hwnd uintptr, callback uintptr, param uintptr) {
	if procHTMLayoutSetCallback == nil {
		return
	}
	procHTMLayoutSetCallback.Call(hwnd, callback, param)
}

func HTMLayoutGetRootElement(hwnd uintptr, handle *uintptr) int {
	if procHTMLayoutGetRootElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetRootElement.Call(hwnd, uintptr(unsafe.Pointer(handle)))
	return int(ret)
}

func HTMLayoutCreateElement(tag *byte, text *uint16, handle *uintptr) int {
	if procHTMLayoutCreateElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutCreateElement.Call(uintptr(unsafe.Pointer(tag)), uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(handle)))
	return int(ret)
}

func HTMLayoutFindElement(hwnd uintptr, pt struct{ X, Y int32 }, handle *uintptr) int {
	if procHTMLayoutFindElement == nil {
		return -1
	}
	ptValue := uintptr(uint32(pt.X)) | (uintptr(uint32(pt.Y)) << 32)
	ret, _, _ := procHTMLayoutFindElement.Call(hwnd, ptValue, uintptr(unsafe.Pointer(handle)))
	return int(ret)
}

func HTMLayoutGetFocusElement(hwnd uintptr, handle *uintptr) int {
	if procHTMLayoutGetFocusElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetFocusElement.Call(hwnd, uintptr(unsafe.Pointer(handle)))
	return int(ret)
}

func HTMLayoutAttachEventHandler(handle uintptr, proc uintptr, tag uintptr) int {
	if procHTMLayoutAttachEventHandler == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutAttachEventHandler.Call(handle, proc, tag)
	return int(ret)
}

func HTMLayoutAttachEventHandlerEx(handle uintptr, proc uintptr, tag uintptr, subscription uint32) int {
	if procHTMLayoutAttachEventHandlerEx == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutAttachEventHandlerEx.Call(handle, proc, tag, uintptr(subscription))
	return int(ret)
}

func HTMLayoutDetachEventHandler(handle uintptr, proc uintptr, tag uintptr) int {
	if procHTMLayoutDetachEventHandler == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutDetachEventHandler.Call(handle, proc, tag)
	return int(ret)
}

func HTMLayoutUpdateElementEx(handle uintptr, flags uint32) int {
	if procHTMLayoutUpdateElementEx == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutUpdateElementEx.Call(handle, uintptr(flags))
	return int(ret)
}

func HTMLayoutSetCapture(handle uintptr) int {
	if procHTMLayoutSetCapture == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetCapture.Call(handle)
	return int(ret)
}

func HTMLayoutShowPopup(popupHandle uintptr, anchorHandle uintptr, placement uint32) int {
	if procHTMLayoutShowPopup == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutShowPopup.Call(popupHandle, anchorHandle, uintptr(placement))
	return int(ret)
}

func HTMLayoutShowPopupAt(popupHandle uintptr, pt struct{ X, Y int32 }, mode uint32) int {
	if procHTMLayoutShowPopupAt == nil {
		return -1
	}
	ptValue := uintptr(uint32(pt.X)) | (uintptr(uint32(pt.Y)) << 32)
	ret, _, _ := procHTMLayoutShowPopupAt.Call(popupHandle, ptValue, uintptr(mode))
	return int(ret)
}

func HTMLayoutHidePopup(popupHandle uintptr) int {
	if procHTMLayoutHidePopup == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutHidePopup.Call(popupHandle)
	return int(ret)
}

func HTMLayoutSelectElements(handle uintptr, selector *byte, callback uintptr, param uintptr) int {
	if procHTMLayoutSelectElements == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSelectElements.Call(handle, uintptr(unsafe.Pointer(selector)), callback, param)
	return int(ret)
}

func HTMLayoutSelectParent(handle uintptr, selector *byte, depth uint32, parent *uintptr) int {
	if procHTMLayoutSelectParent == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSelectParent.Call(handle, uintptr(unsafe.Pointer(selector)), uintptr(depth), uintptr(unsafe.Pointer(parent)))
	return int(ret)
}

func HTMLayoutSendEvent(handle uintptr, eventCode uint32, sourceHandle uintptr, reason uintptr, handled *bool) int {
	if procHTMLayoutSendEvent == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSendEvent.Call(handle, uintptr(eventCode), sourceHandle, reason, uintptr(unsafe.Pointer(handled)))
	return int(ret)
}

func HTMLayoutPostEvent(handle uintptr, eventCode uint32, sourceHandle uintptr, reason uint32) int {
	if procHTMLayoutPostEvent == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutPostEvent.Call(handle, uintptr(eventCode), sourceHandle, uintptr(reason))
	return int(ret)
}

func HTMLayoutGetChildrenCount(handle uintptr, count *uint32) int {
	if procHTMLayoutGetChildrenCount == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetChildrenCount.Call(handle, uintptr(unsafe.Pointer(count)))
	return int(ret)
}

func HTMLayoutGetNthChild(handle uintptr, index uint32, child *uintptr) int {
	if procHTMLayoutGetNthChild == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetNthChild.Call(handle, uintptr(index), uintptr(unsafe.Pointer(child)))
	return int(ret)
}

func HTMLayoutGetElementIndex(handle uintptr, index *int32) int {
	if procHTMLayoutGetElementIndex == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementIndex.Call(handle, uintptr(unsafe.Pointer(index)))
	return int(ret)
}

func HTMLayoutGetParentElement(handle uintptr, parent *uintptr) int {
	if procHTMLayoutGetParentElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetParentElement.Call(handle, uintptr(unsafe.Pointer(parent)))
	return int(ret)
}

func HTMLayoutInsertElement(childHandle uintptr, parentHandle uintptr, index uint32) int {
	if procHTMLayoutInsertElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutInsertElement.Call(childHandle, parentHandle, uintptr(index))
	return int(ret)
}

func HTMLayoutDetachElement(handle uintptr) int {
	if procHTMLayoutDetachElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutDetachElement.Call(handle)
	return int(ret)
}

func HTMLayoutDeleteElement(handle uintptr) int {
	if procHTMLayoutDeleteElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutDeleteElement.Call(handle)
	return int(ret)
}

func HTMLayoutCloneElement(handle uintptr, clone *uintptr) int {
	if procHTMLayoutCloneElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutCloneElement.Call(handle, uintptr(unsafe.Pointer(clone)))
	return int(ret)
}

func HTMLayoutSwapElements(handle1 uintptr, handle2 uintptr) int {
	if procHTMLayoutSwapElements == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSwapElements.Call(handle1, handle2)
	return int(ret)
}

func HTMLayoutSetEventRoot(handle uintptr, prevRoot *uintptr) int {
	if procHTMLayoutSetEventRoot == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetEventRoot.Call(handle, uintptr(unsafe.Pointer(prevRoot)))
	return int(ret)
}

func HTMLayoutScrollToView(handle uintptr, flags uint32) int {
	if procHTMLayoutScrollToView == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutScrollToView.Call(handle, uintptr(flags))
	return int(ret)
}

func HTMLayoutGetElementUID(handle uintptr, uid *uint32) int {
	if procHTMLayoutGetElementUID == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementUID.Call(handle, uintptr(unsafe.Pointer(uid)))
	return int(ret)
}

func HTMLayoutGetElementByUID(hwnd uintptr, uid uint32, handle *uintptr) int {
	if procHTMLayoutGetElementByUID == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementByUID.Call(hwnd, uintptr(uid), uintptr(unsafe.Pointer(handle)))
	return int(ret)
}

func HTMLayoutCallBehaviorMethod(handle uintptr, params *uintptr) int {
	if procHTMLayoutCallBehaviorMethod == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutCallBehaviorMethod.Call(handle, uintptr(unsafe.Pointer(params)))
	return int(ret)
}

func HTMLayoutCombineURL(handle uintptr, buffer *uint16, maxLen uint32) {
	if procHTMLayoutCombineURL == nil {
		return
	}
	procHTMLayoutCombineURL.Call(handle, uintptr(unsafe.Pointer(buffer)), uintptr(maxLen))
}

func HTMLayoutSortElements(handle uintptr, start uint32, end uint32, comparator uintptr, param uintptr) int {
	if procHTMLayoutSortElements == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSortElements.Call(handle, uintptr(start), uintptr(end), comparator, param)
	return int(ret)
}

func HTMLayoutSetTimerEx(handle uintptr, milliseconds uint32, timerId uintptr) int {
	if procHTMLayoutSetTimerEx == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetTimerEx.Call(handle, uintptr(milliseconds), timerId)
	return int(ret)
}

func HTMLayoutGetElementHwnd(handle uintptr, hwnd *uintptr, rootWindow int32) int {
	if procHTMLayoutGetElementHwnd == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementHwnd.Call(handle, uintptr(unsafe.Pointer(hwnd)), uintptr(rootWindow))
	return int(ret)
}

func HTMLayoutGetElementHtml(handle uintptr, html **byte, outer bool) int {
	if procHTMLayoutGetElementHtml == nil {
		return -1
	}
	var outerInt int32
	if outer {
		outerInt = 1
	}
	ret, _, _ := procHTMLayoutGetElementHtml.Call(handle, uintptr(unsafe.Pointer(html)), uintptr(outerInt))
	return int(ret)
}

func HTMLayoutGetElementType(handle uintptr, tag **byte) int {
	if procHTMLayoutGetElementType == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementType.Call(handle, uintptr(unsafe.Pointer(tag)))
	return int(ret)
}

func HTMLayoutSetElementHtml(handle uintptr, html *byte, htmlLength uint32, where uint32) int {
	if procHTMLayoutSetElementHtml == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetElementHtml.Call(handle, uintptr(unsafe.Pointer(html)), uintptr(htmlLength), uintptr(where))
	return int(ret)
}

func HTMLayoutSetElementInnerText(handle uintptr, text *byte, textLength uint32) int {
	if procHTMLayoutSetElementInnerText == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetElementInnerText.Call(handle, uintptr(unsafe.Pointer(text)), uintptr(textLength))
	return int(ret)
}

func HTMLayoutGetElementInnerText(handle uintptr, text **byte) int {
	if procHTMLayoutGetElementInnerText == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementInnerText.Call(handle, uintptr(unsafe.Pointer(text)))
	return int(ret)
}

func HTMLayoutGetAttributeByName(handle uintptr, name *byte, value **uint16) int {
	if procHTMLayoutGetAttributeByName == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetAttributeByName.Call(handle, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(value)))
	return int(ret)
}

func HTMLayoutSetAttributeByName(handle uintptr, name *byte, value *uint16) int {
	if procHTMLayoutSetAttributeByName == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetAttributeByName.Call(handle, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(value)))
	return int(ret)
}

func HTMLayoutGetNthAttribute(handle uintptr, n uint32, name **byte, value **uint16) int {
	if procHTMLayoutGetNthAttribute == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetNthAttribute.Call(handle, uintptr(n), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(value)))
	return int(ret)
}

func HTMLayoutGetAttributeCount(handle uintptr, count *uint32) int {
	if procHTMLayoutGetAttributeCount == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetAttributeCount.Call(handle, uintptr(unsafe.Pointer(count)))
	return int(ret)
}

func HTMLayoutGetStyleAttribute(handle uintptr, name *byte, value **uint16) int {
	if procHTMLayoutGetStyleAttribute == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetStyleAttribute.Call(handle, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(value)))
	return int(ret)
}

func HTMLayoutSetStyleAttribute(handle uintptr, name *byte, value *uint16) int {
	if procHTMLayoutSetStyleAttribute == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutSetStyleAttribute.Call(handle, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(value)))
	return int(ret)
}

func HTMLayoutGetElementState(handle uintptr, state *uint32) int {
	if procHTMLayoutGetElementState == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementState.Call(handle, uintptr(unsafe.Pointer(state)))
	return int(ret)
}

func HTMLayoutSetElementState(handle uintptr, stateToSet uint32, stateToClear uint32, raiseEvent bool) int {
	if procHTMLayoutSetElementState == nil {
		return -1
	}
	var raiseEventInt int32
	if raiseEvent {
		raiseEventInt = 1
	}
	ret, _, _ := procHTMLayoutSetElementState.Call(handle, uintptr(stateToSet), uintptr(stateToClear), uintptr(raiseEventInt))
	return int(ret)
}

func HTMLayoutMoveElement(handle uintptr, x int32, y int32, pos uint32) int {
	if procHTMLayoutMoveElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutMoveElement.Call(handle, uintptr(x), uintptr(y), uintptr(pos))
	return int(ret)
}

func HTMLayoutMoveElementEx(handle uintptr, x int32, y int32, width int32, height int32, pos uint32) int {
	if procHTMLayoutMoveElementEx == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutMoveElementEx.Call(handle, uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(pos))
	return int(ret)
}

func HTMLayoutGetElementLocation(handle uintptr, location *Rect, areas uint32) int {
	if procHTMLayoutGetElementLocation == nil {
		return -1
	}
	ret, _, _ := procHTMLayoutGetElementLocation.Call(handle, uintptr(unsafe.Pointer(location)), uintptr(areas))
	return int(ret)
}

func HTMLayout_UseElement(handle uintptr) int {
	if procHTMLayout_UseElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayout_UseElement.Call(handle)
	return int(ret)
}

func HTMLayout_UnuseElement(handle uintptr) int {
	if procHTMLayout_UnuseElement == nil {
		return -1
	}
	ret, _, _ := procHTMLayout_UnuseElement.Call(handle)
	return int(ret)
}
