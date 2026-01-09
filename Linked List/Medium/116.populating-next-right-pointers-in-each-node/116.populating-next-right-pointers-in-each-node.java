/*
// Definition for a Node.
class Node {
    public int val;
    public Node left;
    public Node right;
    public Node next;

    public Node() {}
    
    public Node(int _val) {
        val = _val;
    }

    public Node(int _val, Node _left, Node _right, Node _next) {
        val = _val;
        left = _left;
        right = _right;
        next = _next;
    }
};
*/

class Solution {
    public Node connect(Node root) {
        Node currentLevelNode=root;
        while(currentLevelNode!=null){
            Node nextLevelStart=null, nextLevelTail=null;
            while(currentLevelNode!=null){
                if(currentLevelNode.left!=null){
                    if(nextLevelStart==null) 
                        nextLevelStart=currentLevelNode.left;
                    else
                        nextLevelTail.next= currentLevelNode.left;
                    nextLevelTail=currentLevelNode.left;
                }
                if(currentLevelNode.right!=null){
                    if(nextLevelStart==null) 
                        nextLevelStart=currentLevelNode.right;
                    else
                        nextLevelTail.next= currentLevelNode.right;
                    nextLevelTail=currentLevelNode.right;
                }
                currentLevelNode = currentLevelNode.next;
            }
            currentLevelNode = nextLevelStart;
        }
        return root;
    }
}