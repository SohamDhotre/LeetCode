/**
 * Definition for a binary tree node.
 * public class TreeNode {
 *     int val;
 *     TreeNode left;
 *     TreeNode right;
 *     TreeNode() {}
 *     TreeNode(int val) { this.val = val; }
 *     TreeNode(int val, TreeNode left, TreeNode right) {
 *         this.val = val;
 *         this.left = left;
 *         this.right = right;
 *     }
 * }
 */
class Solution {
    Map<Integer, Integer>inorderIndexMap;
    int index=0;
    public TreeNode buildTree(int[] preorder, int[] inorder) {
        inorderIndexMap=new HashMap<>(inorder.length);
        for(int i=0;i<inorder.length;i++) inorderIndexMap.put(inorder[i], i);
        return helper(preorder, inorder, 0, inorder.length-1);
    }

    TreeNode helper(int []preorder, int []inorder, int left, int right){
        if(left > right) return null;
        TreeNode node=new TreeNode(preorder[index++]);
        int rootIndex=inorderIndexMap.get(node.val);
        node.left=helper(preorder, inorder, left, rootIndex-1);
        node.right=helper(preorder, inorder, rootIndex+1, right);
        return node;
    }
}