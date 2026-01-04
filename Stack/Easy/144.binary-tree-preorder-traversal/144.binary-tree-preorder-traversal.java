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
    public List<Integer> preorderTraversal(TreeNode root) {
        List<Integer>res=new ArrayList<>();
        pre(root, res);
        return res;
    }

    void pre(TreeNode root, List<Integer>res){
        if(root==null) return;

        res.add(root.val);
        if(root.left!=null) pre(root.left, res);
        if(root.right!=null) pre(root.right, res);
    }
}