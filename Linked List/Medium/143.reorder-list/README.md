# 143. Reorder List

**Difficulty:** Medium  
**Topics:** Linked List, Two Pointers, Stack, Recursion  
**Language:** java

## Problem Link
[LeetCode Problem](https://leetcode.com/problems/reorder-list/)

## Problem Description
You are given the head of a singly linked-list. The list can be represented as:


L0 &rarr; L1 &rarr; &hellip; &rarr; Ln - 1 &rarr; Ln


Reorder the list to be on the following form:


L0 &rarr; Ln &rarr; L1 &rarr; Ln - 1 &rarr; L2 &rarr; Ln - 2 &rarr; &hellip;


You may not modify the values in the list&#39;s nodes. Only nodes themselves may be changed.

 
Example 1:


Input: head = [1,2,3,4]
Output: [1,4,2,3]


Example 2:


Input: head = [1,2,3,4,5]
Output: [1,5,2,4,3]


 
Constraints:


	The number of nodes in the list is in the range [1, 5 * 104].
	1 <= Node.val <= 1000
