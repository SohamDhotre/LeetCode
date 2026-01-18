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
    int count;
    public int goodNodes(TreeNode root) {
        count=0;
        dfs(root, root.val);
        return count;
    }

    void dfs(TreeNode node, int curMax){
        if(node==null) return;
        if(node.val>=curMax){
            curMax=node.val;
            count++;
        } 
        dfs(node.left, curMax);
        dfs(node.right, curMax);
    }
}