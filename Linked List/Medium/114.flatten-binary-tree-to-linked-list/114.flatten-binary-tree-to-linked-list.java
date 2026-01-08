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
    public void flatten(TreeNode root) {
        if(root==null) return;
        helper(root);
    }

    TreeNode helper(TreeNode cur){
        if(cur==null) return cur;
        TreeNode leftTail=helper(cur.left);
        TreeNode rightTail=helper(cur.right);
        if(leftTail!=null){
            leftTail.right=cur.right;
            cur.right=cur.left;
            cur.left=null;
        }
        if (rightTail != null) return rightTail;
        if (leftTail != null) return leftTail;
        return cur;
    }
}