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
    public int kthSmallest(TreeNode root, int k) {
        TreeNode cur=root;
        int count=0;
        while(cur!=null){
            if(cur.left==null){                
                count++;
                if(count==k) return cur.val;
                cur=cur.right;
            } else{
                TreeNode leftSubtreeRightmost=cur.left;
                while(leftSubtreeRightmost.right!=null &&
                       leftSubtreeRightmost.right!=cur){
                    leftSubtreeRightmost = leftSubtreeRightmost.right;
                }
                if(leftSubtreeRightmost.right==null){
                    leftSubtreeRightmost.right=cur;
                    cur=cur.left;
                } else{
                    leftSubtreeRightmost.right=null;
                    count++;
                    if(count==k) return cur.val;
                    cur=cur.right;
                }
            }
        }
        return -1;
    }
}