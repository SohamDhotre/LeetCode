/**
 * Definition for a binary tree node.
 * public class TreeNode {
 *     int val;
 *     TreeNode left;
 *     TreeNode right;
 *     TreeNode(int x) { val = x; }
 * }
 */
class Solution {
    public TreeNode lowestCommonAncestor(TreeNode root, TreeNode p, TreeNode q) {
        if(root==null || root==p || root==q) return root;
        // System.out.println("current root: "+root.val);
        TreeNode left=lowestCommonAncestor(root.left, p, q);
        TreeNode right=lowestCommonAncestor(root.right, p, q);
        // System.out.println("left: "+(left!=null?left.val:"null")+", right: "+(right!=null?right.val:"null"));
        if(left!=null && right!=null) return root;
        return left!=null?left:right;
    }
}