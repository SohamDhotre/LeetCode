# 98. Validate Binary Search Tree

**Difficulty:** Medium  
**Topics:** Tree, Depth-First Search, Binary Search Tree, Binary Tree  
**Language:** java

## Problem Link
[LeetCode Problem](https://leetcode.com/problems/validate-binary-search-tree/)

## Problem Description
Given the root of a binary tree, determine if it is a valid binary search tree (BST).

A valid BST is defined as follows:


	The left subtree of a node contains only nodes with keys strictly less than the node&#39;s key.
	The right subtree of a node contains only nodes with keys strictly greater than the node&#39;s key.
	Both the left and right subtrees must also be binary search trees.


 
Example 1:


Input: root = [2,1,3]
Output: true


Example 2:


Input: root = [5,1,4,null,null,3,6]
Output: false
Explanation: The root node&#39;s value is 5 but its right child&#39;s value is 4.


 
Constraints:


	The number of nodes in the tree is in the range [1, 104].
	-231 <= Node.val <= 231 - 1
