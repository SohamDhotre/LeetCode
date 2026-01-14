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
    Map<Long, Integer>sumFreq;
    int count;
    public int pathSum(TreeNode root, int targetSum) {
        sumFreq=new HashMap<>();
        count=0;
        sumFreq.put(0L, 1);
        dfs(root, 0L, targetSum);
        return count;
    }

    void dfs(TreeNode node, long curSum, int target){
        if(node==null) return;
        curSum += node.val;
        long needed = curSum - target;
        count += sumFreq.getOrDefault(needed, 0);
        sumFreq.put(curSum, sumFreq.getOrDefault(curSum, 0)+1);
        dfs(node.left, curSum, target);
        dfs(node.right, curSum, target);
        sumFreq.put(curSum, sumFreq.get(curSum) - 1);
        if (sumFreq.get(curSum) == 0) sumFreq.remove(curSum);
    }
}