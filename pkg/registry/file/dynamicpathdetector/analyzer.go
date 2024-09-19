package dynamicpathdetector

import (
	"path"
	"strings"
)

func NewPathAnalyzer(threshold int) *PathAnalyzer {
	return &PathAnalyzer{
		RootNodes: make(map[string]*SegmentNode),
		threshold: threshold,
	}
}

func (ua *PathAnalyzer) AnalyzePath(p, identifier string) (string, error) {
	p = path.Clean(p)
	node, exists := ua.RootNodes[identifier]
	if !exists {
		node = &SegmentNode{
			SegmentName: identifier,
			Count:       0,
			Children:    make(map[string]*SegmentNode),
		}
		ua.RootNodes[identifier] = node
	}
	return ua.processSegments(node, p), nil
}

func (ua *PathAnalyzer) processSegments(node *SegmentNode, p string) string {
	var result strings.Builder
	currentNode := node
	start := 0
	for i := range p {
		if p[i] == '/' {
			segment := p[start:i]
			currentNode = ua.processSegment(currentNode, segment)
			ua.updateNodeStats(currentNode)
			result.WriteString(currentNode.SegmentName)
			result.WriteByte('/')
			start = i + 1
		}
	}
	// Process the last segment
	if start < len(p) {
		segment := p[start:]
		currentNode = ua.processSegment(currentNode, segment)
		ua.updateNodeStats(currentNode)
		result.WriteString(currentNode.SegmentName)
	}
	return result.String()
}

func (ua *PathAnalyzer) processSegment(node *SegmentNode, segment string) *SegmentNode {
	if segment == DynamicIdentifier {
		return ua.handleDynamicSegment(node)
	} else if child, exists := node.Children[segment]; exists || node.IsNextDynamic() {
		return ua.handleExistingSegment(node, child, exists)
	} else {
		return ua.handleNewSegment(node, segment)
	}
}

func (ua *PathAnalyzer) handleExistingSegment(node *SegmentNode, child *SegmentNode, exists bool) *SegmentNode {
	if exists {
		return child
	} else {
		return node.Children[DynamicIdentifier]
	}
}

func (ua *PathAnalyzer) handleNewSegment(node *SegmentNode, segment string) *SegmentNode {
	node.Count++
	newNode := &SegmentNode{
		SegmentName: segment,
		Count:       0,
		Children:    make(map[string]*SegmentNode),
	}
	node.Children[segment] = newNode
	return newNode
}

func (ua *PathAnalyzer) handleDynamicSegment(node *SegmentNode) *SegmentNode {
	if dynamicChild, exists := node.Children[DynamicIdentifier]; exists {
		return dynamicChild
	} else {
		return ua.createDynamicNode(node)
	}
}

func (ua *PathAnalyzer) createDynamicNode(node *SegmentNode) *SegmentNode {
	dynamicNode := &SegmentNode{
		SegmentName: DynamicIdentifier,
		Count:       0,
		Children:    make(map[string]*SegmentNode),
	}

	// Copy all existing children to the new dynamic node
	for _, child := range node.Children {
		shallowChildrenCopy(child, dynamicNode)
	}

	// Replace all children with the new dynamic node
	node.Children = map[string]*SegmentNode{
		DynamicIdentifier: dynamicNode,
	}

	return dynamicNode
}

func (ua *PathAnalyzer) updateNodeStats(node *SegmentNode) {
	if node.Count > ua.threshold && !node.IsNextDynamic() {
		dynamicChild := &SegmentNode{
			SegmentName: DynamicIdentifier,
			Count:       0,
			Children:    make(map[string]*SegmentNode),
		}

		// Copy all descendants
		for _, child := range node.Children {
			shallowChildrenCopy(child, dynamicChild)
		}

		node.Children = map[string]*SegmentNode{
			DynamicIdentifier: dynamicChild,
		}
	}
}

func shallowChildrenCopy(src, dst *SegmentNode) {
	for segmentName := range src.Children {
		if _, ok := dst.Children[segmentName]; !ok {
			dst.Children[segmentName] = src.Children[segmentName]
		} else {
			dst.Children[segmentName].Count += src.Children[segmentName].Count
			shallowChildrenCopy(src.Children[segmentName], dst.Children[segmentName])
		}
	}
}

func CompareDynamic(dynamicPath, regularPath string) bool {
	dynamicIndex, regularIndex := 0, 0
	dynamicLen, regularLen := len(dynamicPath), len(regularPath)

	for dynamicIndex < dynamicLen && regularIndex < regularLen {
		// Find the next segment in dynamicPath
		dynamicSegmentStart := dynamicIndex
		for dynamicIndex < dynamicLen && dynamicPath[dynamicIndex] != '/' {
			dynamicIndex++
		}
		dynamicSegment := dynamicPath[dynamicSegmentStart:dynamicIndex]

		// Find the next segment in regularPath
		regularSegmentStart := regularIndex
		for regularIndex < regularLen && regularPath[regularIndex] != '/' {
			regularIndex++
		}
		regularSegment := regularPath[regularSegmentStart:regularIndex]

		if dynamicSegment != DynamicIdentifier && dynamicSegment != regularSegment {
			return false
		}

		// Move to the next segment
		dynamicIndex++
		regularIndex++
	}

	return true
}
